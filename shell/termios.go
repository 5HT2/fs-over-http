// really google, you couldnt provide this?
package main

/*
#include <termios.h>
*/
import "C";
import "runtime";

const SYS_IOCTL uintptr = 53;

type tcflag_t uint32;
type cc_t     byte;

const ICANON    tcflag_t = C.ICANON;

const TCSANOW   int = 0;
const TCSADRAIN int = 1;
const TCSAFLUSH int = 2;

const NCCS byte = 32;

type termios struct {
	c_iflag    tcflag_t; /* input modes */
	c_oflag    tcflag_t; /* output modes */
	c_cflag    tcflag_t; /* control modes */
	c_lflag    tcflag_t; /* local modes */
	c_cc[NCCS] cc_t;     /* special characters*/
};

// technically fd should be an int but golang
func tcgetattr(fd uintptr, termios_p *termios) (int) {
	if (runtime.GOOS == "linux") {
		var hi C.struct_termios;

		C.tcgetattr(C.int(fd), &hi);

		termios_p.c_iflag = tcflag_t(hi.c_iflag);
		termios_p.c_oflag = tcflag_t(hi.c_oflag);
		termios_p.c_cflag = tcflag_t(hi.c_cflag);
		termios_p.c_lflag = tcflag_t(hi.c_lflag);
		
		for i := 0; i < int(NCCS); i++ {
			termios_p.c_cc[i] = cc_t(hi.c_cc[i]);
		}

		return 0;
	} else { return -1; }
}

func tcsetattr(fd uintptr, act int, termios_p *termios) (int) {
	if (runtime.GOOS == "linux") {
		var hi C.struct_termios;

		hi.c_iflag = C.uint(termios_p.c_iflag);
		hi.c_oflag = C.uint(termios_p.c_oflag);
		hi.c_cflag = C.uint(termios_p.c_cflag);
		hi.c_lflag = C.uint(termios_p.c_lflag);

		for i := 0; i < int(NCCS); i++ {
			hi.c_cc[i] = C.uchar(termios_p.c_cc[i]);
		}

		C.tcsetattr(C.int(fd), C.int(act), &hi);

		return 0;
	} else { return -1; }
}