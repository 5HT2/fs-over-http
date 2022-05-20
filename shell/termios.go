package main;

type tcflag_t uint32;
type cc_t     byte;

const TCSANOW   = 0;
const TCSADRAIN = 1;
const TCSAFLUSH = 2;

const NCCS byte = 32;

type termios struct {
	c_iflag    tcflag_t; /* input modes */
	c_oflag    tcflag_t; /* output modes */
	c_cflag    tcflag_t; /* control modes */
	c_lflag    tcflag_t; /* local modes */
	c_cc[NCCS] cc_t;     /* special characters*/
};

// technically fd should be an int but golang
func tcgetattr(fd uintptr, termios_p *termios) {
	// TODO
}

func tcsetattr(fd uintptr, act int, termios_p *termios) {
	// TODO
}