package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/natefinch/lumberjack"
)

//Config file
//local = Path to folder containing "latest" folder, used as a temporary store. The .zip used will be placed into the latest folder along with the executable running
//remote = Path to remote folder containing the versioned zip files
//runexe = the name of the executable to run, multiple could exist within the archive

//Run, - "go run ."
//Build - "go build -o Tool_Network_Updater.exe ."

func main() {
	// Setup log rotation
	// Save log file to ProgramData
	programData := os.Getenv("ProgramData")
	logPath := programData + "\\UpdaterTool\\updater.log"
	// Ensure directory exists
	_ = os.MkdirAll(programData+"\\UpdaterTool", 0755)
	log.SetOutput(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    5, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Compress:   true,
	})

	// CLI flags
	local := flag.String("local", "", "Path to local folder containing 'latest'")
	remote := flag.String("remote", "", "Path to remote folder containing versioned zip files")
	runexe := flag.String("runexe", "", "Name of the executable to run")
	flag.Parse()

	if *local == "" || *remote == "" || *runexe == "" {
		log.Fatal("All flags --local, --remote, --runexe are required.")
	}

	pathexe := *local + "\\latest\\" + *runexe

	log.Println("Starting updater...")
	log.Printf("Remote folder: %s", *remote)
	log.Printf("Local folder: %s", *local)
	log.Printf("Executable to run: %s", *runexe)

	file, currVersRemote, forced, err := readLatestFromFolder(*remote, true, "Remote")

	if err != nil {
		log.Printf("Error reading remote folder: %v", err)
		log.Printf("Attempting to launch previous executable: %s", pathexe)
		launchApp(pathexe)
		os.Exit(0)
	}

	log.Printf("Remote version: %v, Forced update: %v", currVersRemote, forced)
	_, currVersLocal, _, err := readLatestFromFolder(*local, false, "Local")
	log.Printf("Local version: %v", currVersLocal)

	if currVersRemote > currVersLocal || forced {
		log.Printf("Local: %v", currVersLocal)
		log.Printf("Remote: %v", currVersRemote)
		log.Println("New version found, updating...")
		//Copy to local path
		log.Printf("Copying %s from remote to local...", file)
		dst, err := os.Create(*local + "\\" + file)
		check(err)
		src, err := os.Open(*remote + "\\" + file)
		check(err)
		_, err = io.Copy(dst, src)
		check(err)

		//Clear latest folder
		log.Println("Clearing latest folder...")
		err = os.RemoveAll(*local + "\\latest")
		check(err)

		src.Close()
		dst.Close()

		//Unzip to local location
		log.Printf("Unzipping %s to latest folder...", file)
		updateLatest(*local+"\\"+file, *local+"\\latest")

		//Copy zip into latest, root will be deleted
		log.Printf("Copying zip %s into latest folder...", file)
		move, _ := os.Create(*local + "\\latest\\" + file)
		src, _ = os.Open(*local + "\\" + file)
		_, err = io.Copy(move, src)
		check(err)

		move.Close()
		src.Close()

		//Delete latest local zip copy
		log.Printf("Deleting local zip copy: %s", *local+"\\"+file)
		err = os.RemoveAll(*local + "\\" + file)
		check(err)
	} else {
		log.Println("No update required. Local version is up to date.")
	}

	if _, err = os.Stat(pathexe); os.IsNotExist(err) {
		log.Printf("Executable not found in latest folder. Unzipping again...")
		updateLatest(*local+"\\"+file, *local+"\\latest")
	}

	log.Printf("Launching application: %s", pathexe)
	launchApp(pathexe)
	log.Println("Updater finished successfully.")
	// Wait for user input before exiting
	log.Println("Press Enter to exit...")
	_, _ = os.Stdin.Read(make([]byte, 1))
}

func launchApp(path string) {
	log.Printf("Launching: %s", path)
	cmnd := exec.Command(path)
	err := cmnd.Start()
	check(err)
}
