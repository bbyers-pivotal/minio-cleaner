package main

import (
	"flag"
	"fmt"
	"github.com/minio/minio-go"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type minioFileInfo struct {
	filename string
	date     time.Time
}

type minioFile []minioFileInfo

func (s minioFile) Less(i, j int) bool                { return s[i].date.Before(s[j].date) }
func (s minioFile) Swap(i, j int)                     { s[i], s[j] = s[j], s[i] }
func (s minioFile) Len() int                          { return len(s) }
func (s minioFile) ToDelete(keep int) []minioFileInfo { return s[keep:len(s)] }
func (s minioFile) ToKeep(keep int) []minioFileInfo   { return s[0:keep] }

func main() {
	endpoint := flag.String("host", "", "Enter Minio host")
	accessKeyID := flag.String("access-key", "", "Enter access key")
	secretAccessKey := flag.String("secret-key", "", "Enter secret key")
	prefix := flag.String("prefix", "", "Enter prefix for the files to delete")
	useSSL := flag.Bool("use-ssl", true, "Use SSL for Minio connection?")
	dryRun := flag.Bool("dry-run", true, "Set whether is dry run or will delete")
	bucket := flag.String("bucket", "", "Enter name of bucket to use")
	numberOfBackupsToKeep := flag.Int("backups-to-keep", 9999, "Enter number of backups to keep")

	flag.Parse()

	if len(*endpoint) == 0 {
		fmt.Println("No endpoint specified, please set -host to the host you want to use")
		os.Exit(1)
	}

	if len(*accessKeyID) == 0 {
		fmt.Println("No access key specified, please set -access-key to your access key")
		os.Exit(1)
	}

	if len(*secretAccessKey) == 0 {
		fmt.Println("No secret key specified, please set -secret-key to your access key")
		os.Exit(1)
	}

	if len(*prefix) == 0 {
		fmt.Println("No prefix specified, please set -prefix to the prefix of the files to delete")
		os.Exit(1)
	}

	if len(*bucket) == 0 {
		fmt.Println("No bucket specified, please set -bucket to the bucket you want to use")
		os.Exit(1)
	}

	// Initialize minio client object.
	minioClient, err := minio.New(*endpoint, *accessKeyID, *secretAccessKey, *useSSL)
	if err != nil {
		log.Fatalln(err)
	}

	exists, err := minioClient.BucketExists(*bucket)

	if !exists {
		fmt.Println("Minio bucket does not exist")
		fmt.Println(err)
		os.Exit(1)
	}

	isRecursive := true

	//before/after size of the prefix'd items -v verbose flag

	doneCh := make(chan struct{})
	defer close(doneCh)

	minioFiles := []minioFileInfo{}

	objectCh := minioClient.ListObjectsV2(*bucket, *prefix, isRecursive, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return
		}

		//setup regex to grab the date pieces from the filenames
		r, _ := regexp.Compile(`\d{4}_\d{2}_\d{2}_\d{2}_\d{2}_\d{2}`)

		//grab the date portion of the filename
		stringMatch := r.FindSubmatch([]byte(object.Key))

		//grab the parts that consist of the date and the time so we can format it properly
		dateTimeStringMatch := string(stringMatch[0])
		dateChunkFormatted := strings.Replace(dateTimeStringMatch[0:10], "_", "-", -1)
		timeChunkFormatted := strings.Replace(dateTimeStringMatch[11:len(dateTimeStringMatch)], "_", ":", -1)

		//build a new string with the date/time pieces in a parsable format
		var dateTimeSB strings.Builder
		dateTimeSB.WriteString(dateChunkFormatted)
		dateTimeSB.WriteString(" ")
		dateTimeSB.WriteString(timeChunkFormatted)
		dateTimeString := dateTimeSB.String()

		dateTime, _ := time.Parse("2006-01-02 15:04:05", dateTimeString)

		//create a new file object we can sort through and be able to delete
		minioFiles = append(minioFiles, minioFileInfo{date: dateTime, filename: object.Key})
	}

	var files minioFile = minioFiles

	//sort files newest first
	sort.Sort(sort.Reverse(files))

	if files.Len() >= *numberOfBackupsToKeep {
		toDelete := files.ToDelete(*numberOfBackupsToKeep)

		if len(toDelete) > 0 {
			for _, m := range (toDelete) {
				if (*dryRun) {
					fmt.Println("DRY RUN - File " + m.filename + " would be deleted")
				} else {
					fmt.Println("Deleting", m.filename)
					minioClient.RemoveObject(*bucket, m.filename)
				}
			}
		}
	} else {
		fmt.Println("No backups to prune")
	}
}
