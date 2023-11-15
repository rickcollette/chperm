package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

var (
	versionFlag bool
	verbose     = flag.Bool("vvv", false, "enable verbose output")
	pathFlag    = flag.String("path", "", "path to apply permissions to")
	permsFlag   = flag.String("perms", "", "permissions to apply (in octal format)")
	recurseFlag = flag.Bool("recurse", false, "recurse into directories")
	rollbackNum = flag.Int("rollback", 0, "number of changes to rollback")
	audit       = flag.Bool("audit", false, "enable audit logging")
	outputType  = flag.String("o", "csv", "output format (csv, excel)")
	filename    string
	saveFile    string
)

const (
	instPath     = "/usr/local"
	rollbackPath = instPath + "/share/rollback/"
	appVersion   = ""
)

func init() {
	// Generate a unique filename when the application starts
	filename = "audit_" + time.Now().Format("20060102_5")

	flag.BoolVar(&versionFlag, "version", false, "print version and exit")

	// Check for -v flag manually
	for _, arg := range os.Args {
		if arg == "-v" {
			versionFlag = true
			break
		}
	}
}
func main() {
	flag.Usage = usage
	flag.Parse()
	if versionFlag {
		printVersion()
	}
	if *audit {
		if outputType == nil || (*outputType != "csv" && *outputType != "excel") {
			fmt.Println("Invalid output format:", *outputType)
			return
		}
	}
	// Handle rollback request
	if *rollbackNum > 0 {
		rollbackChanges(*rollbackNum)
		return
	}

	if *pathFlag != "" && *permsFlag != "" {
		permissions, err := permissionsOct(*permsFlag)
		if err != nil {
			fmt.Println("Invalid permissions format:", err)
			return
		}

		if *recurseFlag {
			applyPermissionsRecursively(*pathFlag, permissions)
		} else {
			applyPermissions(*pathFlag, permissions)
		}
	} else {
		processConfigFile()
	}
}

func processConfigFile() {
	configFile := instPath + "/etc/chperm.conf"
	fmt.Print("configFile: ", configFile, "\n")

	file, err := os.Open(configFile)
	if err != nil {
		fmt.Printf("Error opening configuration file: %v\\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if skipLine(line) {
			continue
		}

		path, permissions, recurse, err := parseLine(line)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if *verbose {
			fmt.Printf("Processing: %s with permissions %04o\\n", path, permissions)
		}

		if recurse {
			applyPermissionsRecursively(path, permissions)
		} else {
			applyPermissions(path, permissions)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading configuration file: %v\\n", err)
	}
}

func skipLine(line string) bool {
	return strings.HasPrefix(line, "#") || len(line) == 0
}

func parseLine(line string) (string, os.FileMode, bool, error) {
	parts := strings.Split(line, ",")
	if len(parts) != 3 {
		return "", 0, false, fmt.Errorf("invalid line format: %s", line)
	}

	path := os.ExpandEnv(strings.TrimSpace(parts[0]))
	permissions, err := permissionsOct(strings.TrimSpace(parts[1]))
	if err != nil {
		return "", 0, false, fmt.Errorf("invalid permissions: %s", parts[1])
	}

	recurse := strings.TrimSpace(parts[2]) == "1"
	return path, permissions, recurse, nil
}

func permissionsOct(perm string) (os.FileMode, error) {
	oct, err := strconv.ParseUint(perm, 8, 32)
	if err != nil {
		return 0, err
	}
	return os.FileMode(oct), nil
}

func printVersion() {
	fmt.Println("chperm version:", appVersion)
	os.Exit(0)
}

func applyPermissions(path string, permissions os.FileMode) {
	matches, err := filepath.Glob(path)
	if err != nil {
		fmt.Printf("Error matching path: %v\\n", err)
		return
	}

	for _, match := range matches {
		if _, err := os.Stat(match); os.IsNotExist(err) {
			if *verbose {
				fmt.Printf("Path does not exist: %s\\n", match)
			}
			continue
		}

		oldPerms, _ := getPermissions(match)
		if *verbose {
			fmt.Printf("Setting permissions for %s to %04o\\n", match, permissions)
		}

		err := os.Chmod(match, permissions)
		if err != nil {
			fmt.Printf("Error setting permissions for %s: %v\\n", match, err)
		} else {
			logChangeAudit(getCurrentUser(), path, oldPerms, permissions, time.Now())
			logChange(match, oldPerms, permissions)
		}
	}
}

func applyPermissionsRecursively(path string, permissions os.FileMode) {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if *verbose {
				fmt.Printf("Error accessing path %q: %v\n", path, err)
			}
			return err
		}

		oldPerms := info.Mode().Perm()
		if err := os.Chmod(path, permissions); err != nil {
			fmt.Printf("Error setting permissions for %s: %v\n", path, err)
		} else {
			logChangeAudit(getCurrentUser(), path, oldPerms, permissions, time.Now())
			logChange(path, oldPerms, permissions)
			if *verbose {
				fmt.Printf("Changed permissions for %s from %04o to %04o\n", path, oldPerms, permissions)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
	}
}

func logChange(path string, oldPerms, newPerms os.FileMode) {
	// Ensure the rollback directory exists
	os.MkdirAll(rollbackPath, 0755)

	// Create a log file with a timestamp
	timestamp := time.Now().Format("20060102-5")
	logFileName := filepath.Join(rollbackPath, fmt.Sprintf("chperm-%s.log", timestamp))

	// Format the log entry
	logEntry := fmt.Sprintf("%s,%04o,%04o\n", path, oldPerms, newPerms)

	// Append log entry to the log file
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(logEntry); err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
	}
}

func rollbackChanges(n int) {
	files, err := os.ReadDir(rollbackPath)
	if err != nil {
		fmt.Println("Error reading rollback directory:", err)
		return
	}

	// Sort files by name (timestamp) in reverse order
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() > files[j].Name()
	})

	// Process the last 'n' files for rollback
	for i := 0; i < n && i < len(files); i++ {
		rollbackFile(filepath.Join(rollbackPath, files[i].Name()))
	}
}

func getPermissions(path string) (os.FileMode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Mode().Perm(), nil
}

func rollbackFile(filepath string) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println("Error reading rollback file:", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 3 {
			fmt.Println("Invalid entry in rollback file:", line)
			continue
		}
		filePath, oldPermsStr := parts[0], parts[1]
		oldPerms, err := permissionsOct(oldPermsStr)
		if err != nil {
			fmt.Println("Invalid permissions format in rollback file:", err)
			continue
		}
		if err := os.Chmod(filePath, oldPerms); err != nil {
			fmt.Printf("Error rolling back permissions for %s: %v\n", filePath, err)
		} else if *verbose {
			fmt.Printf("Rolled back permissions for %s to %04o\n", filePath, oldPerms)
		}
	}
}

func getCurrentUser() string {
	user, err := user.Current()
	if err != nil {
		return "unknown"
	}
	return user.Username
}

func logChangeAudit(who, filePath string, oldPerms, newPerms os.FileMode, timestamp time.Time) {
	if *audit {
		// Implement logic based on output type
		switch *outputType {
		case "csv":
			writeToCSV(who, filePath, oldPerms, newPerms, timestamp)
		case "excel":
			writeToExcel(who, filePath, oldPerms, newPerms, timestamp)
		default:
			fmt.Println("Unsupported output format:", *outputType)
		}
	}
}

func writeToCSV(who, filePath string, oldPerms, newPerms os.FileMode, timestamp time.Time) {
	saveFile = filename + ".csv"
	file, err := os.OpenFile(saveFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening CSV file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{
		who,
		filePath,
		fmt.Sprintf("%04o", oldPerms),
		fmt.Sprintf("%04o", newPerms),
		timestamp.Format(time.RFC3339),
	}

	if err := writer.Write(record); err != nil {
		fmt.Println("Error writing to CSV file:", err)
	}
}

func writeToExcel(who, filePath string, oldPerms, newPerms os.FileMode, timestamp time.Time) {
	saveFile = filename + ".xlsx"
	f, err := excelize.OpenFile(saveFile)
	if err != nil {
		// Create a new file if it doesn't exist
		f = excelize.NewFile()
		// Create a default sheet
		f.NewSheet("Sheet1")
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("Error closing Excel file:", err)
		}
	}()

	sheetName := "Sheet1"

	// Find the next empty row in the sheet
	rowIndex := getNextEmptyRow(f, sheetName)

	// Write data to the specified row
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), who)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIndex), filePath)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIndex), strconv.FormatUint(uint64(oldPerms), 8))
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIndex), strconv.FormatUint(uint64(newPerms), 8))
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowIndex), timestamp.Format(time.RFC3339))

	// Save the file
	if err := f.SaveAs(filename); err != nil {
		fmt.Println("Error saving Excel file:", err)
	}
}

func getNextEmptyRow(f *excelize.File, sheetName string) int {
	row := 1
	for {
		cell, _ := f.GetCellValue(sheetName, fmt.Sprintf("A%d", row))
		if cell == "" {
			break
		}
		row++
	}
	return row
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Println("  Apply permissions recursively:")
	fmt.Println("  ", os.Args[0], "-path /path/to/folder -perms 0755 -recurse")
	fmt.Println("  Apply permissions without recursion:")
	fmt.Println("  ", os.Args[0], "-path /path/to/file -perms 0644")
	fmt.Println("  Rollback last 5 permission changes:")
	fmt.Println("  ", os.Args[0], "-rollback 5")
	fmt.Println("  Audit changes to permissions and output csv:")
	fmt.Println("  ", os.Args[0], "-audit -o csv")
	fmt.Println("  Audit changes to permissions and output excel:")
	fmt.Println("  ", os.Args[0], "-audit -o excel")
	fmt.Println("\nPermission Bits:")
	fmt.Println("  Permissions in Unix are represented by three groups:")
	fmt.Println("  - Owner permissions")
	fmt.Println("  - Group permissions")
	fmt.Println("  - Others' permissions")
	fmt.Println("  Each group can have read (r), write (w), and execute (x) permissions.")
	fmt.Println("  Represented in octal format, these permissions are:")
	fmt.Println("  - Read (r) is 4")
	fmt.Println("  - Write (w) is 2")
	fmt.Println("  - Execute (x) is 1")
	fmt.Println("  To combine permissions, add the values together. For example:")
	fmt.Println("  - Read and write (rw) is 6 (4+2)")
	fmt.Println("  - Read, write, and execute (rwx) is 7 (4+2+1)")
}
