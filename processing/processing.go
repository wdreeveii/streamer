	var outputfile *os.File
	outputfile, err = os.OpenFile("test_VIDEO", os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer outputfile.Close()

	cmd := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "avi", "pipe:1")
	//cmd := exec.Command("cat")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	cmd.Stdout = outputfile
	err = cmd.Start()
	if err != nil {
		return err
	}