package cmd

import (
	"fmt"
	"os"

	"github.com/ckeyer/tarofs/pkgs/fs"
	"github.com/ckeyer/tarofs/pkgs/storage/levelfs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(MoundCmd())
}

func MoundCmd() *cobra.Command {
	var (
		mountDir string
		leveldir string
	)

	cmd := &cobra.Command{
		Use:     "mount",
		Aliases: []string{"m"},
		Short:   "mount tarofs to a directory.",
		PreRun: func(cmd *cobra.Command, args []string) {
			logrus.SetFormatter(&logrus.JSONFormatter{})
			if err := checkDir(mountDir); err != nil {
				logrus.Fatalf("check mountDir faield, %s", err)
			}
			if err := checkDir(leveldir); err != nil {
				logrus.Fatalf("check leveldir faield, %s", err)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			stgr, err := levelfs.NewLevelStorage(leveldir)
			if err != nil {
				logrus.Fatalf("new levelfs storage failed, %s", err)
				return
			}
			filesys, err := fs.NewFS(mountDir, stgr, stgr)
			if err != nil {
				logrus.Fatal("new mount falied, ", err)
			}
			// c, err := fs.Mount(mountDir)

			// defer c.Close()
			// logrus.Infof("mount %s successful.", mountDir)

			go waitExec(func() {
				if err := filesys.Close(); err != nil {
					logrus.Fatalf("umount %s failed, %s", mountDir, err)
				}
				logrus.Fatalf("umount %s successful.", mountDir)
			})

			// if p := c.Protocol(); !p.HasInvalidate() {
			// 	logrus.Fatalf("kernel FUSE support is too old to have invalidations: version %v", p)
			// }

			if err := filesys.Serve(); err != nil {
				logrus.Fatal("start file system serve failed, ", err)
			}

		},
	}

	cmd.Flags().StringVarP(&mountDir, "mount-point", "m", "/tmp/tarofs", "mount point directory.")
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
