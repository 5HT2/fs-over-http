// im so sorry
package main;

import "strings";
import "bufio";
import "fmt";
import "os";
import "github.com/valyala/fasthttp";
import "unicode/utf8";
import "mime/multipart";
import "bytes";
import "io";

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

func pstatus(status int) (bool) {
	switch (status) {
		case 200:		
			fmt.Printf("\n200 OK\n");
			return false;
		case 204:
			fmt.Printf("\n204 NO CONTENT\n");
			return false;	
		case 403:
			fmt.Printf("\n403 FORBIDDEN\n");
		case 404:
			fmt.Printf("\n404 NOT FOUND\n");
		case 500:
			fmt.Printf("\n500 INTERNAL SERVER ERROR\n")
		default:
			fmt.Printf("\n%3d UNKNOWN\n", status);
	}

	return true;
}

func test_runes(data []byte, str string, off int) (bool) {
	var pos1 int = off;
	var pos2 int = 0;

	for (pos2 < len(str)) {
		r1, s := utf8.DecodeRuneInString(str[pos2:]);
		r2, _ := utf8.DecodeRune(data[pos1:]);

		if (r1 != r2) {
			return false;
		}

		pos1 += s;
		pos2 += s;
	}

	return true;
}

func parse_directory(data []byte) (string, []string, bool) {
	var valid bool = false;
	var pos   int  = 0;
	
	var name       string = "";
	var contents []string;

	for pos < len(data) {
		r, s := utf8.DecodeRune(data[pos:]);

		name += string(r);
		pos  += s;

		if (test_runes(data, "/\n", pos)) {
			pos += len("/\n");
			valid = true;
			break;
		}
	}

	if (!valid) {
		return "", nil, false;
	}

	for pos < len(data) {
		r, s := utf8.DecodeRune(data[pos:]);
		pos += s;

		if (r == '├' || r == '└') {
			var last bool = r == '└';

			if (test_runes(data, "── ", pos)) {
				pos += len("── ");

				var file string = "";

				valid = false;

				for pos < len(data) {
					r, s = utf8.DecodeRune(data[pos:]);
					pos += s;

					if (r == '\n') {
						valid = true;
						break;
					}

					file += string(r);
				}

				if (!valid) {
					return "", nil, false;
				}

				contents = append(contents, file);
			} else {
				return "", nil, false;
			}

			if (last) {
				break;
			}
		} else {
			return "", nil, false;
		}
	}

	return name, contents, true;
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

	var token     string;
	var htoken    bool   = false;
	
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
						var fillin, compl = do_completions(cmd, []string{"exit", "connect", "ls", "cat", "get", "put", "mkdir", "rm"});
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

								var req fasthttp.Request;
								var res fasthttp.Response;

								req.SetRequestURI(largv[1]);
								
								if (len(largv) > 2) {
									req.Header.Set("Auth", largv[2]);
								}

								err := fasthttp.Do(&req, &res);

								if (err != nil) {
									fmt.Printf("\nXXX %s\n", err.Error());
									goto cmddone;
								}

								if (pstatus(res.StatusCode())) {
									goto cmddone;
								}

								if (len(largv) > 2) {
									htoken = true;
									token  = largv[2];
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
									if (htoken) {
										req.Header.Set("Auth", token);
									}

									err := fasthttp.Do(&req, &res);

									if (err != nil) {
										fmt.Printf("\nXXX %s\n", err.Error());
										goto cmddone;
									}

									if (pstatus(res.StatusCode())) {
										goto cmddone;
									} else if (res.StatusCode() == 200) {
										var data []byte = res.Body();
										
										dir_name, dir, is_dir := parse_directory(data);

										if (is_dir) {
											fmt.Printf("%s:\n", dir_name);
											
											for i := 0; i < len(dir); i++ {
												fmt.Printf("%s ", dir[i]);
											}

											fmt.Printf("\n");
										} else {
											fmt.Printf("XXX not a directory\n");
										}
									}
 								} else {
									fmt.Printf("\nXXX not connected\n");
								}
							} else {
								fmt.Printf("\nXXX path not specified\n")
							}

							goto cmddone; 
						} else if (largv[0] == "cat") {
							if (connected) {
								if (len(largv) > 1) {
									var req fasthttp.Request;
									var res fasthttp.Response;

									req.SetRequestURI(server + largv[1]);
									if (htoken) {
										req.Header.Set("Auth", token);
									}

									err := fasthttp.Do(&req, &res);

									if (err != nil) {
										fmt.Printf("\nXXX %s\n", err.Error());
										goto cmddone;
									}

									var data []byte = res.Body();

									_, _, is_dir := parse_directory(data);

									if (!is_dir) {
										if (pstatus(res.StatusCode())) {
											goto cmddone;
										} else if (res.StatusCode() == 200) {
											fmt.Printf("%s\n", string(data));
										}
									} else {
										fmt.Printf("\nXXX is a directory\n");
									}
								} else {
									fmt.Printf("\nXXX path not specified\n")
								}
							} else {
								fmt.Printf("\nXXX not connected\n");
							}

							goto cmddone;
						} else if (largv[0] == "dog") {
							fmt.Printf("\nwoof!\n");
							
							goto cmddone;
						} else if (largv[0] == "get") {
							if (connected) {
								if (len(largv) > 1) {
									if (len(largv) > 2) {
										var req fasthttp.Request;
										var res fasthttp.Response;

										req.SetRequestURI(server + largv[1]);
										if (htoken) {
											req.Header.Set("Auth", token);
										}

										err = fasthttp.Do(&req, &res);

										if (err != nil) {
											fmt.Printf("\nXXX %s\n", err.Error());
											goto cmddone;
										}

										var data []byte = res.Body();

										_, _, is_dir := parse_directory(data);

										if (!is_dir) {
											if (pstatus(res.StatusCode())) {
												goto cmddone;
											} else if (res.StatusCode() == 200) {
												fd, err := os.OpenFile(largv[2], os.O_WRONLY | os.O_CREATE | os.O_EXCL, 0600);

												if (err != nil) {
													fmt.Printf("\nXXX %s\n", err.Error());
													goto cmddone;
												}

												_, err = fd.Write(data)

												if (err != nil) {
													fmt.Printf("\nXXX %s\n", err.Error());
												}

												fd.Close();
											} else if (res.StatusCode() == 204) {
												fd, err := os.OpenFile(largv[2], os.O_WRONLY | os.O_CREATE | os.O_EXCL, 0600);

												if (err != nil) {
													fmt.Printf("\nXXX %s\n", err.Error());
													goto cmddone;
												}

												fd.Close();
											}
										} else {
											fmt.Printf("\nXXX is a directory\n");
										}
									} else {
										fmt.Printf("\nXXX local path not specified\n");
									}
								} else {
									fmt.Printf("\nXXX remote path not specified\n");
								}
							} else {
								fmt.Printf("\nXXX not connected\n");
							}

							goto cmddone;
						} else if (largv[0] == "put") {
							if (connected) {
								if (len(largv) > 1) {
									if (len(largv) > 2) {
										fd, err := os.OpenFile(largv[1], os.O_RDONLY, 0);

										if (err != nil) {
											fmt.Printf("\nXXX %s\n", err.Error());
											fd.Close();
											goto cmddone;
										}

										var fbuf    bytes.Buffer;
										var writer *multipart.Writer = multipart.NewWriter(&fbuf);

										fwriter, err := writer.CreateFormFile("file", fd.Name());										
										if (err != nil) {
											fmt.Printf("\nXXX %s\n", err.Error());
											fd.Close();
											goto cmddone;
										}

										_, err = io.Copy(fwriter, fd);
										if (err != nil) {
											fmt.Printf("\nXXX %s\n", err.Error());
											fd.Close();
											goto cmddone;
										}

										fd.Close();
										writer.Close();

										var req fasthttp.Request;
										var res fasthttp.Response;

										req.SetRequestURI(server + largv[2]);
										req.Header.SetMethod("POST");
										req.Header.SetContentType(writer.FormDataContentType());
										req.SetBody(fbuf.Bytes());
										if (htoken) {
											req.Header.Set("Auth", token);
										}

										err = fasthttp.Do(&req, &res);

										if (err != nil) {
											fmt.Printf("\nXXX %s\n", err.Error());
											goto cmddone;
										}

										pstatus(res.StatusCode());
									} else {
										fmt.Printf("\nXXX remote path not specified\n");
									}
								} else {
									fmt.Printf("\nXXX local path not specified\n")
								}
							} else {
								fmt.Printf("\nXXX not connected\n");
							}

							goto cmddone;
						} else if (largv[0] == "mkdir") {
							if (connected) {
								if (len(largv) > 1) {
									var fbuf    bytes.Buffer;
									var writer *multipart.Writer = multipart.NewWriter(&fbuf);

									fwriter, err := writer.CreateFormField("dir");
									if (err != nil) {
										fmt.Printf("\nXXX %s\n", err.Error());
										goto cmddone;
									}

									_, err = io.WriteString(fwriter, largv[1]);
									if (err != nil) {
										fmt.Printf("\nXXX %s\n", err.Error());
										goto cmddone;
									}

									writer.Close();

									var req fasthttp.Request;
									var res fasthttp.Response;

									req.SetRequestURI(server + largv[1]);
									req.Header.SetMethod("POST");
									req.Header.SetContentType(writer.FormDataContentType());
									req.SetBody(fbuf.Bytes());
									if (htoken) {
										req.Header.Set("Auth", token);
									}

									err = fasthttp.Do(&req, &res);

									if (err != nil) {
										fmt.Printf("\nXXX %s\n", err.Error());
										goto cmddone;
									}

									pstatus(res.StatusCode());
								} else {
									fmt.Printf("\nXXX remote path not specified\n");
								}
							} else {
								fmt.Printf("\nXXX not connected\n");
							}

							goto cmddone;
						} else if (largv[0] == "rm") {
							if (connected) {
								if (len(largv) > 1) {
									var req fasthttp.Request;
									var res fasthttp.Response;

									req.SetRequestURI(server + largv[1]);
									req.Header.SetMethod("DELETE");
									if (htoken) {
										req.Header.Set("Auth", token);
									}

									err = fasthttp.Do(&req, &res);

									if (err != nil) {
										fmt.Printf("\nXXX %s\n", err.Error());
										goto cmddone;
									}

									pstatus(res.StatusCode());
								} else {
									fmt.Printf("\nXXX remote path not specified\n");
								}
							} else {
								fmt.Printf("\nXXX not connected\n");
							}

							goto cmddone;
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