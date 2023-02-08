package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func getQuerryTimeout() int {
	fmt.Println("How long should queries run before logging? (in seconds)")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	number, _ := strconv.Atoi(strings.TrimSpace(input))
	return number
}

func credentialCheck() (credentials string) {
	out, err := exec.Command("rpm", "-qa").Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	output := string(out)
	if strings.Contains(output, "psa-") {
		fmt.Println("Plesk System Found")
		shadow, err := exec.Command("cat", "/etc/psa/.psa.shadow").Output()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		credentials = "-uadmin -p" + strings.TrimSpace(string(shadow))
	} else if strings.Contains(output, "cpanel") {
		fmt.Println("WHM System Found")
		credentials = "-uroot"
	} else {
		fmt.Println("Enter MySQL root password:")
		reader := bufio.NewReader(os.Stdin)
		password, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		credentials = "-uroot -p" + strings.TrimSpace(password)
	}
	return credentials
}

func createDirectories() {
	exec.Command("mkdir", "-p", "/var/log/mysql").Run()
	exec.Command("touch", "/var/log/mysql/slow-queries.log").Run()
	exec.Command("chown", "-R", "mysql:mysql", "/var/log/mysql").Run()
}

func enableSlowQueryLogging(credentials string, number int) {
	fmt.Println("Enabling Slow Queries")

	db, err := sqlx.Open("mysql", credentials+"@tcp(localhost:3306)/")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	_, err = db.Exec("SET GLOBAL slow_query_log = ON")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Setting long query time to", number)

	_, err = db.Exec("SET GLOBAL long_query_time = ?", number)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Setting to log queries not using indexes")

	_, err = db.Exec("SET GLOBAL log_queries_not_using_indexes = ON")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Setting log file location to /var/log/mysql/slow-queries.log")

	_, err = db.Exec("SET GLOBAL slow_query_log_file='/var/log/mysql/slow-queries.log'")
	if err != nil {
		fmt.Println(err)
		return
	}
}

func main() {
	number := getQuerryTimeout()
	credentials := credentialCheck()
	createDirectories()
	enableSlowQueryLogging(credentials, number)
	fmt.Println("\nComplete")
}
