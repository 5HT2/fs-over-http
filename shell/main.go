package main;

import "bufio";
import "fmt";
import "os";

func main() {
	var stdin *bufio.Reader = bufio.NewReader(os.Stdin);

	var orig_tios termios;
	var nbuf_tios termios;

	tcgetattr(os.Stdin.Fd(), &orig_tios);

	nbuf_tios          =  orig_tios;

	nbuf_tios.c_lflag &= ^ICANON;

	println(nbuf_tios.c_lflag);

	tcsetattr(os.Stdin.Fd(), TCSANOW, &nbuf_tios);

	for (true) { 
		fmt.Printf("fs-over-http $ ");

		for (true) {
			var bt, err = stdin.ReadByte()

			if (err == nil) {
				fmt.Printf("%i", bt);
			}
		}

		break;
	}

	tcsetattr(os.Stdin.Fd(), TCSANOW, &orig_tios);
}