package globpath

/*
BSD 2-Clause License

Copyright (c) 2017, Karrick McDermott
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"unsafe"
)

// DefaultScratchBufferSize specifies the size of the scratch buffer that will
// be allocated by Walk, ReadDirents, or ReadDirnames when a scratch buffer is
// not provided or the scratch buffer that is provided is smaller than
// MinimumScratchBufferSize bytes. This may seem like a large value; however,
// when a program intends to enumerate large directories, having a larger
// scratch buffer results in fewer operating system calls.
const DefaultScratchBufferSize = 64 * 1024

// MinimumScratchBufferSize specifies the minimum size of the scratch buffer
// that Walk, ReadDirents, and ReadDirnames will use when reading file entries
// from the operating system. It is initialized to the result from calling
// `os.Getpagesize()` during program startup.
var MinimumScratchBufferSize int

func init() {
	MinimumScratchBufferSize = os.Getpagesize()
}

// Options provide parameters for how the Walk function operates.
type Options struct {
	// ErrorCallback specifies a function to be invoked in the case of an error
	// that could potentially be ignored while walking a file system
	// hierarchy. When set to nil or left as its zero-value, any error condition
	// causes Walk to immediately return the error describing what took
	// place. When non-nil, this user supplied function is invoked with the OS
	// pathname of the file system object that caused the error along with the
	// error that took place. The return value of the supplied ErrorCallback
	// function determines whether the error will cause Walk to halt immediately
	// as it would were no ErrorCallback value provided, or skip this file
	// system node yet continue on with the remaining nodes in the file system
	// hierarchy.
	//
	// ErrorCallback is invoked both for errors that are returned by the
	// runtime, and for errors returned by other user supplied callback
	// functions.
	ErrorCallback func(string, error) ErrorAction

	// FollowSymbolicLinks specifies whether Walk will follow symbolic links
	// that refer to directories. When set to false or left as its zero-value,
	// Walk will still invoke the callback function with symbolic link nodes,
	// but if the symbolic link refers to a directory, it will not recurse on
	// that directory. When set to true, Walk will recurse on symbolic links
	// that refer to a directory.
	FollowSymbolicLinks bool

	// Unsorted controls whether or not Walk will sort the immediate descendants
	// of a directory by their relative names prior to visiting each of those
	// entries.
	//
	// When set to false or left at its zero-value, Walk will get the list of
	// immediate descendants of a particular directory, sort that list by
	// lexical order of their names, and then visit each node in the list in
	// sorted order. This will cause Walk to always traverse the same directory
	// tree in the same order, however may be inefficient for directories with
	// many immediate descendants.
	//
	// When set to true, Walk skips sorting the list of immediate descendants
	// for a directory, and simply visits each node in the order the operating
	// system enumerated them. This will be more fast, but with the side effect
	// that the traversal order may be different from one invocation to the
	// next.
	Unsorted bool

	// Callback is a required function that Walk will invoke for every file
	// system node it encounters.
	Callback WalkFunc

	// PostChildrenCallback is an option function that Walk will invoke for
	// every file system directory it encounters after its children have been
	// processed.
	PostChildrenCallback WalkFunc

	// ScratchBuffer is an optional byte slice to use as a scratch buffer for
	// Walk to use when reading directory entries, to reduce amount of garbage
	// generation. Not all architectures take advantage of the scratch
	// buffer. If omitted or the provided buffer has fewer bytes than
	// MinimumScratchBufferSize, then a buffer with DefaultScratchBufferSize
	// bytes will be created and used once per Walk invocation.
	ScratchBuffer []byte
}

// ErrorAction defines a set of actions the Walk function could take based on
// the occurrence of an error while walking the file system. See the
// documentation for the ErrorCallback field of the Options structure for more
// information.
type ErrorAction int

const (
	// Halt is the ErrorAction return value when the upstream code wants to halt
	// the walk process when a runtime error takes place. It matches the default
	// action the Walk function would take were no ErrorCallback provided.
	Halt ErrorAction = iota

	// SkipNode is the ErrorAction return value when the upstream code wants to
	// ignore the runtime error for the current file system node, skip
	// processing of the node that caused the error, and continue walking the
	// file system hierarchy with the remaining nodes.
	SkipNode
)

// WalkFunc is the type of the function called for each file system node visited
// by Walk. The pathname argument will contain the argument to Walk as a prefix;
// that is, if Walk is called with "dir", which is a directory containing the
// file "a", the provided WalkFunc will be invoked with the argument "dir/a",
// using the correct os.PathSeparator for the Go Operating System architecture,
// GOOS. The directory entry argument is a pointer to a Dirent for the node,
// providing access to both the basename and the mode type of the file system
// node.
//
// If an error is returned by the Callback or PostChildrenCallback functions,
// and no ErrorCallback function is provided, processing stops. If an
// ErrorCallback function is provided, then it is invoked with the OS pathname
// of the node that caused the error along along with the error. The return
// value of the ErrorCallback function determines whether to halt processing, or
// skip this node and continue processing remaining file system nodes.
//
// The exception is when the function returns the special value
// filepath.SkipDir. If the function returns filepath.SkipDir when invoked on a
// directory, Walk skips the directory's contents entirely. If the function
// returns filepath.SkipDir when invoked on a non-directory file system node,
// Walk skips the remaining files in the containing directory. Note that any
// supplied ErrorCallback function is not invoked with filepath.SkipDir when the
// Callback or PostChildrenCallback functions return that special value.
type WalkFunc func(osPathname string, directoryEntry *Dirent) error

// Walk walks the file tree rooted at the specified directory, calling the
// specified callback function for each file system node in the tree, including
// root, symbolic links, and other node types. The nodes are walked in lexical
// order, which makes the output deterministic but means that for very large
// directories this function can be inefficient.
//
// This function is often much faster than filepath.Walk because it does not
// invoke os.Stat for every node it encounters, but rather obtains the file
// system node type when it reads the parent directory.
//
// If a runtime error occurs, either from the operating system or from the
// upstream Callback or PostChildrenCallback functions, processing typically
// halts. However, when an ErrorCallback function is provided in the provided
// Options structure, that function is invoked with the error along with the OS
// pathname of the file system node that caused the error. The ErrorCallback
// function's return value determines the action that Walk will then take.
//
//    func main() {
//        dirname := "."
//        if len(os.Args) > 1 {
//            dirname = os.Args[1]
//        }
//        err := Walk(dirname, &Options{
//            Callback: func(osPathname string, de *Dirent) error {
//                fmt.Printf("%s %s\n", de.ModeType(), osPathname)
//                return nil
//            },
//            ErrorCallback: func(osPathname string, err error) ErrorAction {
//            	// Your program may want to log the error somehow.
//            	fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
//
//            	// For the purposes of this example, a simple SkipNode will suffice,
//            	// although in reality perhaps additional logic might be called for.
//            	return SkipNode
//            },
//        })
//        if err != nil {
//            fmt.Fprintf(os.Stderr, "%s\n", err)
//            os.Exit(1)
//        }
//    }
func Walk(pathname string, options *Options) error {
	pathname = filepath.Clean(pathname)

	var fi os.FileInfo
	var err error

	if options.FollowSymbolicLinks {
		fi, err = os.Stat(pathname)
		if err != nil {
			return err
		}
	} else {
		fi, err = os.Lstat(pathname)
		if err != nil {
			return err
		}
	}

	mode := fi.Mode()
	if mode&os.ModeDir == 0 {
		return fmt.Errorf("cannot Walk non-directory: %s", pathname)
	}

	dirent := &Dirent{
		name:     filepath.Base(pathname),
		modeType: mode & os.ModeType,
	}

	// If ErrorCallback is nil, set to a default value that halts the walk
	// process on all operating system errors. This is done to allow error
	// handling to be more succinct in the walk code.
	if options.ErrorCallback == nil {
		options.ErrorCallback = defaultErrorCallback
	}

	if len(options.ScratchBuffer) < MinimumScratchBufferSize {
		options.ScratchBuffer = make([]byte, DefaultScratchBufferSize)
	}

	err = walk(pathname, dirent, options)
	if err == filepath.SkipDir {
		return nil // silence SkipDir for top level
	}
	return err
}

// defaultErrorCallback always returns Halt because if the upstream code did not
// provide an ErrorCallback function, walking the file system hierarchy ought to
// halt upon any operating system error.
func defaultErrorCallback(_ string, _ error) ErrorAction { return Halt }

// walk recursively traverses the file system node specified by pathname and the
// Dirent.
func walk(osPathname string, dirent *Dirent, options *Options) error {
	err := options.Callback(osPathname, dirent)
	if err != nil {
		if err == filepath.SkipDir {
			return err
		}
		err = errCallback(err.Error()) // wrap potential errors returned by callback
		if action := options.ErrorCallback(osPathname, err); action == SkipNode {
			return nil
		}
		return err
	}

	// On some platforms, an entry can have more than one mode type bit set.
	// For instance, it could have both the symlink bit and the directory bit
	// set indicating it's a symlink to a directory.
	if dirent.IsSymlink() {
		if !options.FollowSymbolicLinks {
			return nil
		}
		skip, err := symlinkDirHelper(osPathname, dirent, options)
		if err != nil || skip {
			return err
		}
	}

	if !dirent.IsDir() {
		return nil
	}

	// If get here, then specified pathname refers to a directory.
	deChildren, err := readdirents(osPathname, options.ScratchBuffer)
	if err != nil {
		if action := options.ErrorCallback(osPathname, err); action == SkipNode {
			return nil
		}
		return err
	}

	if !options.Unsorted {
		sort.Sort(deChildren) // sort children entries unless upstream says to leave unsorted
	}

	for _, deChild := range deChildren {
		osChildname := filepath.Join(osPathname, deChild.name)
		err = walk(osChildname, deChild, options)
		if err != nil {
			if err != filepath.SkipDir {
				return err
			}
			// If received skipdir on a directory, stop processing that
			// directory, but continue to its siblings. If received skipdir on a
			// non-directory, stop processing remaining siblings.
			if deChild.IsSymlink() {
				var skip bool
				if skip, err = symlinkDirHelper(osChildname, deChild, options); err != nil {
					return err
				}
				if skip {
					continue
				}
			}
			if !deChild.IsDir() {
				// If not directory, return immediately, thus skipping remainder
				// of siblings.
				return nil
			}
		}
	}

	if options.PostChildrenCallback == nil {
		return nil
	}

	err = options.PostChildrenCallback(osPathname, dirent)
	if err == nil || err == filepath.SkipDir {
		return err
	}

	err = errCallback(err.Error()) // wrap potential errors returned by callback
	if action := options.ErrorCallback(osPathname, err); action == SkipNode {
		return nil
	}
	return err
}

// On unix we don't need to fixup pathnames with symlinks when doing a Stat().
func evalSymlinksHelper(pathname string) (string, error) {
	return pathname, nil
}

// symlinkDirHelper accepts a pathname and Dirent for a symlink, and determine
// what the ModeType bits are for what the symlink ultimately points to (such
// as whether it is a directory or not)
func symlinkDirHelper(pathname string, de *Dirent, o *Options) (bool, error) {
	// Only need to Stat entry if platform did not already have os.ModeDir
	// set, such as would be the case for unix like operating systems.
	// (This guard eliminates extra os.Stat check on Windows.)
	if !de.IsDir() {

		fi, err := os.Stat(pathname)
		if err != nil {
			if action := o.ErrorCallback(pathname, err); action == SkipNode {
				return true, nil
			}
			return false, err
		}
		de.modeType = fi.Mode() & os.ModeType
	}
	return false, nil
}

// returns a sortable slice of pointers to Dirent structures, each
// representing the file system name and mode type for one of the immediate
// descendant of the specified directory. If the specified directory is a
// symbolic link, it will be resolved.
//
// If an optional scratch buffer is provided that is at least one page of
// memory, it will be used when reading directory entries from the file system.
//
//    children, err := readdirents(osDirname, nil)
//    if err != nil {
//        return nil, errors.Wrap(err, "cannot get list of directory children")
//    }
//    sort.Sort(children)
//    for _, child := range children {
//        fmt.Printf("%s %s\n", child.ModeType, child.Name)
//    }
func readdirents(osDirname string, scratchBuffer []byte) (Dirents, error) {
	dh, err := os.Open(osDirname)
	if err != nil {
		return nil, err
	}

	var entries Dirents

	fd := int(dh.Fd())

	if len(scratchBuffer) < MinimumScratchBufferSize {
		scratchBuffer = make([]byte, DefaultScratchBufferSize)
	}

	var de *syscall.Dirent

	for {
		n, err := syscall.ReadDirent(fd, scratchBuffer)
		if err != nil {
			_ = dh.Close() // ignore potential error returned by Close
			return nil, err
		}
		if n <= 0 {
			break // end of directory reached
		}
		// Loop over the bytes returned by reading the directory entries.
		buf := scratchBuffer[:n]
		for len(buf) > 0 {
			de = (*syscall.Dirent)(unsafe.Pointer(&buf[0])) // point entry to first syscall.Dirent in buffer
			buf = buf[de.Reclen:]                           // advance buffer

			if de.Ino == 0 {
				continue // this item has been deleted, but not yet removed from directory
			}

			nameSlice := nameFromDirent(de)
			namlen := len(nameSlice)
			if (namlen == 0) || (namlen == 1 && nameSlice[0] == '.') || (namlen == 2 && nameSlice[0] == '.' && nameSlice[1] == '.') {
				continue // skip unimportant entries
			}
			osChildname := string(nameSlice)

			// Convert syscall constant, which is in purview of OS, to a
			// constant defined by Go, assumed by this project to be stable.
			var mode os.FileMode
			switch de.Type {
			case syscall.DT_REG:
				// regular file
			case syscall.DT_DIR:
				mode = os.ModeDir
			case syscall.DT_LNK:
				mode = os.ModeSymlink
			case syscall.DT_CHR:
				mode = os.ModeDevice | os.ModeCharDevice
			case syscall.DT_BLK:
				mode = os.ModeDevice
			case syscall.DT_FIFO:
				mode = os.ModeNamedPipe
			case syscall.DT_SOCK:
				mode = os.ModeSocket
			default:
				// If syscall returned unknown type (e.g., DT_UNKNOWN, DT_WHT),
				// then resolve actual mode by getting stat.
				fi, err := os.Lstat(filepath.Join(osDirname, osChildname))
				if err != nil {
					_ = dh.Close() // ignore potential error returned by Close
					return nil, err
				}
				// We only care about the bits that identify the type of a file
				// system node, and can ignore append, exclusive, temporary,
				// setuid, setgid, permission bits, and sticky bits, which are
				// coincident to the bits that declare type of the file system
				// node.
				mode = fi.Mode() & os.ModeType
			}

			entries = append(entries, &Dirent{name: osChildname, modeType: mode})
		}
	}
	if err = dh.Close(); err != nil {
		return nil, err
	}
	return entries, nil
}

type errCallback string

func (e errCallback) Error() string { return string(e) }
