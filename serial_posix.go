// +build !windows

package serial

// #include <termios.h>
// #include <unistd.h>
import "C"

// TODO: Maybe change to using syscall package + ioctl instead of cgo

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	//"unsafe"
)

// The arguments to Flush()
const (
	TCIFLUSH  int = int(C.TCIFLUSH)
	TCOFLUSH  int = int(C.TCOFLUSH)
	TCIOFLUSH int = int(C.TCIOFLUSH)
)

func openPort(name string, baud int) (rwc io.ReadWriteCloser, err error, fdesc uintptr) {
	f, err := os.OpenFile(name, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0666)
	if err != nil {
		return
	}

	fd := C.int(f.Fd())
	if C.isatty(fd) != 1 {
		f.Close()
		return nil, errors.New("File is not a tty"), 0
	}

	var st C.struct_termios
	_, err = C.tcgetattr(fd, &st)
	if err != nil {
		f.Close()
		return nil, err, 0
	}
	var speed C.speed_t
	switch baud {
	case 115200:
		speed = C.B115200
	case 57600:
		speed = C.B57600
	case 38400:
		speed = C.B38400
	case 19200:
		speed = C.B19200
	case 9600:
		speed = C.B9600
	case 4800:
		speed = C.B4800
	case 2400:
		speed = C.B2400
	default:
		f.Close()
		return nil, fmt.Errorf("Unknown baud rate %v", baud), 0
	}

	_, err = C.cfsetispeed(&st, speed)
	if err != nil {
		f.Close()
		return nil, err, 0
	}
	_, err = C.cfsetospeed(&st, speed)
	if err != nil {
		f.Close()
		return nil, err, 0
	}

	// Select local mode
	st.c_cflag |= (C.CLOCAL | C.CREAD)

	// Select raw mode
	st.c_lflag &= ^C.tcflag_t(C.ICANON | C.ECHO | C.ECHOE | C.ISIG)
	st.c_oflag &= ^C.tcflag_t(C.OPOST)

	_, err = C.tcsetattr(fd, C.TCSANOW, &st)
	if err != nil {
		f.Close()
		return nil, err, 0
	}

	//fmt.Println("Tweaking", name)
	r1, _, e := syscall.Syscall(syscall.SYS_FCNTL,
		uintptr(f.Fd()),
		uintptr(syscall.F_SETFL),
		uintptr(0))
	if e != 0 || r1 != 0 {
		s := fmt.Sprint("Clearing NONBLOCK syscall error:", e, r1)
		f.Close()
		return nil, errors.New(s), 0
	}

	/*
				r1, _, e = syscall.Syscall(syscall.SYS_IOCTL,
			                uintptr(f.Fd()),
			                uintptr(0x80045402), // IOSSIOSPEED
			                uintptr(unsafe.Pointer(&baud)));
			        if e != 0 || r1 != 0 {
			                s := fmt.Sprint("Baudrate syscall error:", e, r1)
					f.Close()
		                        return nil, os.NewError(s)
				}
	*/

	return f, nil, f.Fd()
}

func Flush(fd uintptr, queue int) (err error) {
	_, err = C.tcflush(C.int(fd), C.int(queue))
	return
}
