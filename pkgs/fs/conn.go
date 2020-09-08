package fs

import (
	"bazil.org/fuse"
)

// Mount .
func Mount(mountpoint string) (*fuse.Conn, error) {
	return fuse.Mount(
		mountpoint,
		fuse.FSName("tarofs"),
		fuse.VolumeName("Taro File System"),

		fuse.LocalVolume(),

		fuse.NoAppleDouble(),
		fuse.NoAppleXattr(),

		fuse.ExclCreate(),
		fuse.DaemonTimeout("3600"),
		fuse.AllowOther(),
		fuse.AllowSUID(),

		// fuse.DefaultPermissions(),
		// fuse.MaxReadahead(1024*128), // TODO: not tested yet, possibly improving read performance
		fuse.AsyncRead(),
		fuse.WritebackCache(),
	)
}

// Umount .
func Umount(mountpoint string) error {
	return fuse.Unmount(mountpoint)
}
