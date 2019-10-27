package cmd

import (
	"fmt"
	"os"

	"github.com/ckeyer/tarofs/pkgs/levelfs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
)

func init() {
	rootCmd.AddCommand(MoundCmd())
}

func MoundCmd() *cobra.Command {
	var (
		mountpoint string
		leveldir   string
	)

	cmd := &cobra.Command{
		Use:     "mount",
		Aliases: []string{"m"},
		Short:   "mount tarofs to a directory.",
		PreRun: func(cmd *cobra.Command, args []string) {
			logrus.SetFormatter(&logrus.JSONFormatter{})
			if err := checkDir(mountpoint); err != nil {
				logrus.Fatalf("check mountpoint faield, %s", err)
			}
			if err := checkDir(leveldir); err != nil {
				logrus.Fatalf("check leveldir faield, %s", err)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			db, err := leveldb.OpenFile(leveldir, nil)
			if err != nil {
				logrus.Fatalf("open leveldb failed, %s", err)
			}

			c, err := levelfs.Mount(mountpoint)
			if err != nil {
				logrus.Fatal("mount falied, ", err)
			}
			defer c.Close()
			logrus.Infof("mount %s successful.", mountpoint)

			go waitExec(func() {
				if err := levelfs.Umount(mountpoint); err != nil {
					logrus.Fatalf("umount %s failed, %s", mountpoint, err)
				}
				logrus.Fatalf("umount %s successful.", mountpoint)
			})

			if p := c.Protocol(); !p.HasInvalidate() {
				logrus.Fatalf("kernel FUSE support is too old to have invalidations: version %v", p)
			}

			filesys := levelfs.NewFS(c, db)
			if err := filesys.Serve(); err != nil {
				logrus.Fatal("start file system serve failed, ", err)
			}

			// Check if the mount process has an error to report.
			<-c.Ready
			if err := c.MountError; err != nil {
				logrus.Fatal("mount file system failed, ", err)
			}
		},
	}

	cmd.Flags().StringVarP(&mountpoint, "mount-point", "m", "/tmp/tarofs", "mount point directory.")
	cmd.Flags().StringVarP(&leveldir, "leveldb-dir", "l", "/data/tarofs_data", "leveldb data directory.")
	return cmd
}

func checkDir(dir string) error {
	fi, err := os.Stat(dir)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("open %s failed, %s", dir, err)
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("mkdir %s failed, %s", dir, err)
		}
		logrus.Debugf("mkdir %s", dir)
	} else if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory.", dir)
	}

	return nil
}
