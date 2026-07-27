package notifier

import "flag"

func main() {
}

func configPath() string {
	filePath := flag.String("configPath", "", "configuration file")
	flag.Parse()
	return *filePath
}
