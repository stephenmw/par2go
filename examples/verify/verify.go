package main

import (
	"fmt"
	"github.com/stephenmw/par2go/par2"
	"log"
	"os"
	"runtime"
)

var NUM_WORKERS = runtime.NumCPU()

type job struct {
	file      *par2.File
	corrupted []int
}

func init() {
	runtime.GOMAXPROCS(NUM_WORKERS)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Not enough arguments")
	}

	parfile, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	rs := new(par2.RecoverySet)
	errs := rs.ReadRecoveryFile(parfile)
	fmt.Println(errs)

	files := par2.FilesSortedByName(rs.Files)

	assign := make(chan job)
	result := make(chan job)

	for i := 0; i < NUM_WORKERS; i++ {
		go worker(rs, assign, result)
	}

	num_jobs := len(files)
	for i := 0; i < len(files); i++ {
		select {
		case j := <-result:
			fmt.Printf("%s: %v\n", j.file.Name, j.corrupted)
			num_jobs--
			i--
		case assign <- job{file: files[i]}:
		}
	}

	close(assign)

	for num_jobs > 0 {
		j := <-result
		fmt.Printf("%s: %v\n", j.file.Name, j.corrupted)
		num_jobs--
	}

	/*
		for _, f := range files {
			fp, err := os.Open(f.Name)
			if err != nil {
				log.Fatal(err)
			}

			f.Fp = fp
			indexes, _ := rs.CheckFile(f.Id)
			fmt.Printf("%s: %v\n", f.Name, indexes)
		}
	*/
}

func worker(rs *par2.RecoverySet, in <-chan job, out chan<- job) {
	for j := range in {
		fp, err := os.Open(j.file.Name)
		if err != nil {
			log.Fatal(err)
		}

		j.file.Fp = fp
		indexes, _ := rs.CheckFile(j.file.Id)
		j.corrupted = indexes

		out <- j
	}
}
