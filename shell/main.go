package main;

import "strings";
import "bufio";
import "fmt";
import "os";
import "github.com/valyala/fasthttp";
import "unicode/utf8";

func do_completions(cmd string, cmps []string) (bool, string) {
	var filt []string;

	for i := 0; i < len(cmps); i++ {
		if (strings.HasPrefix(cmps[i], cmd)) {
			filt = append(filt, cmps[i]);
		}
	}
	
	if (len(filt) == 1) {
		return true, filt[0];
	} else if (len(filt) > 1) {
		fmt.Printf("\n");
		for i := 0; i < len(filt); i++ {
			fmt.Printf("%s ", filt[i]);
		}
		fmt.Print("\n");
	}

	return false, "";
}

func main() {
	var stdin *bufio.Reader = bufio.NewReader(os.Stdin);

	var orig_tios termios;
	var nbuf_tios termios;

	tcgetattr(os.Stdin.Fd(), &orig_tios);

	nbuf_tios          =  orig_tios;
	nbuf_tios.c_lflag &= ^(ICANON | ISIG | ECHO);

	tcsetattr(os.Stdin.Fd(), TCSANOW, &nbuf_tios);

	var server    string;
	var connected bool   = false;
	for (true) { 
		fmt.Printf("fs-over-http $ ");

		var cpos int = 0;
		var cmd  string;
		for (true) {
			var bt, err = stdin.ReadByte();

			if (err == nil) {
				if (bt == 0x03) { // ctrl + c
					goto exit;
				} else if (bt == 0x09) { // tab
					var largv []string = strings.Split(cmd, " ");

					if (len(largv) == 1) {
						var fillin, compl = do_completions(cmd, []string{"exit", "connect"});
						if (fillin) {
							largv[0] = compl;
							cmd = strings.Join(largv, " ");
							cpos = len(largv[0]);
						}
					}

					fmt.Printf("\x1b[2K\x1b[1Gfs-over-http $ %s\x1b[%dG", cmd, cpos + len("fs-over-http $ ") + 1);
				} else if (bt == 0x0a) {
					var largv []string = strings.Split(cmd, " ");

					if (len(largv) > 0) {
						if (largv[0] == "exit") {
							goto exit;
						} else if (largv[0] == "connect") {
							if (len(largv) > 1) {
								if (!strings.HasPrefix(largv[1], "http://")) {
									largv[1] = "http://" + largv[1];
								}

								if (!(len(strings.Split(largv[1], ":")) > 2)) {
									largv[1] = largv[1] + ":6060";
								}

								status, _, err := fasthttp.Get([]byte{}, largv[1]);

								if (err != nil || status != 200) {
									fmt.Printf("\nERROR %d %s\n", status, err.Error());
									goto cmddone;
								}

								server    = largv[1] + "/";
								connected = true;

								fmt.Printf("\n");
								
								goto cmddone;
							}
						} else if (largv[0] == "ls") {
							if (len(largv) > 1) {
								if (connected) {
									var req fasthttp.Request;
									var res fasthttp.Response;

									req.SetRequestURI(server + largv[1]);

									err := fasthttp.Do(&req, &res);

									if (err != nil) {
										fmt.Printf("\nERROR %d %s\n", res.StatusCode(), err.Error());
										goto cmddone;
									}

									if (res.StatusCode() == 200) {
										fmt.Printf("\n200 OK\n");
										if (string(res.Header.ContentType()) == "application/x-directory") {
											var data []byte = res.Body();

											var dir string;
											var i   int;

											for i = 0; i < len(data) && data[i] != '\n'; i++ {
												dir = dir + string(data[i]);
											}

											fmt.Printf("%s:\n", dir);
											i++;

											for (i < len(data)) {
												r, size := utf8.DecodeRune(data[i:]);
												i += size;

												if (r == '├' || r == '└') {
													var name   string;

													i += len("── ");
													
													for (data[i] != '\n') {
														name += string(data[i]);
														i++;
													}

													i++;
													
													fmt.Printf("%s ", name);

													if (r == '└') {
														break;
													}
												}
											}
											
											fmt.Printf("\n");
											goto cmddone;
										} else {
											fmt.Printf("ERROR XXX not a directory\n");
											goto cmddone;
										}
									}
								} else {
									fmt.Printf("\nERROR XXX not connected\n");
									goto cmddone;
								}
							}
						}
					}

					// cmderr:
					fmt.Printf("\ninvalid command \"%s\"\nfs-over-http $ ", cmd);

					cmddone:
					cmd  = "";
					cpos = 0;
					fmt.Printf("\x1b[2K\x1b[1Gfs-over-http $ %s\x1b[%dG", cmd, cpos + len("fs-over-http $ ") + 1);
				} else if (bt == 0x1b) {
					bt, err = stdin.ReadByte();

					if (bt == '[') {
						bt, err = stdin.ReadByte();

						if (bt == 'C') { // move cursor right
							if (cpos < len(cmd)) {
								cpos++;
								fmt.Printf("\x1b[2K\x1b[1Gfs-over-http $ %s\x1b[%dG", cmd, cpos + len("fs-over-http $ ") + 1);
							}
						} else if (bt == 'D') { // move cursor left
							if (cpos > 0) {
								fmt.Printf("\x1b[2K\x1b[1Gfs-over-http $ %s\x1b[%dG", cmd, cpos + len("fs-over-http $ "));
								cpos--;
							}
						}
					}
				} else if (bt == 0x7f) { // backspace
					if (cpos > 0) {
						cmd = cmd[:cpos - 1] + cmd[cpos:];
						cpos--;
						fmt.Printf("\x1b[2K\x1b[1Gfs-over-http $ %s\x1b[%dG", cmd, cpos + len("fs-over-http $ ") + 1);	
					}
				} else {
					if (cpos < len(cmd)) {
						cmd = cmd[:cpos] + string(bt) + cmd[cpos:];
					} else {
						cmd += string(bt);
					}

					cpos++;
					fmt.Printf("\x1b[2K\x1b[1Gfs-over-http $ %s\x1b[%dG", cmd, cpos + len("fs-over-http $ ") + 1);
				}
			}
		}

		break;
	}

	exit:
	fmt.Printf("\n");
	tcsetattr(os.Stdin.Fd(), TCSANOW, &orig_tios);
}