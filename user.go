package user

import (
        "errors"
        "fmt"
        "strings"
        "os"
        "os/user"
        "io/ioutil"
        "path/filepath"
        "strconv"
)

// The User definition
type User struct {
        Username         string
        Uid              string
        Gid              string
        Groupname        string
        Homedir          string
        Shell            string
        GroupMemberships []string
}

// Exported function for creating new users
func (u *User) AddUser() error {

        err := u.PreFlightChecks()
        if err != nil {
                return err
        }

        passwdLine := u.passwdLine()
        err = appendFile("/etc/passwd", passwdLine)
        if err != nil {
                return err
        }

        groupLine := u.groupLine()
        err = appendFile("/etc/group", groupLine)
        if err != nil {
                return err
        }

        err = u.addAdditionalGroups()
        if err != nil {
                return err
        }

        err = u.makeHomeDir()
        if err != nil {
                return err
        }

        err = u.recursiveChownHome()
        if err != nil {
                return err
        }

        return nil

}

// Create the users home directory
func (u *User) makeHomeDir() error {

        if _, err := os.Stat(u.Homedir); os.IsNotExist(err) {
                err := os.Mkdir(u.Homedir, 0755)
                if err != nil {
                        return err
                }
        }

        return nil

}

// Append user to extra groups to grant user additional group memberships
func (u *User) addAdditionalGroups() error {
        input, err := ioutil.ReadFile("/etc/group")
        if err != nil {
                return err
        }

        lines := strings.Split(string(input), "\n")

        for i, line := range lines {

                group := strings.Split(line, ":")[0]
                for _, extraGroup := range u.GroupMemberships {
                        if group == extraGroup {
                                if len(strings.Split(line, ":")[3]) == 0 {
                                        lines[i] = fmt.Sprintf("%s%s", line, u.Username)
                                } else {
                                        lines[i] = fmt.Sprintf("%s,%s", line, u.Username)
                                }
                        }

                }

        }

        output := strings.Join(lines, "\n")

        fmt.Println(output)

        err = ioutil.WriteFile("/etc/group", []byte(output), 0644)
        if err != nil {
                return err
        }

        return nil
}

// Append a line to important files
func appendFile(filename string, line string) error {

        f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
        if err != nil {
                return err
        }
        defer f.Close()
        if _, err := f.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
                return err
        }

        return nil

}

// Build up a /etc/passwd file line
func (u *User) passwdLine() string {

        return fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s", u.Username, u.Groupname, u.Uid, u.Gid, "", u.Homedir, u.Shell)

}

// Build up a /etc/group file line
func (u *User) groupLine() string {

        return fmt.Sprintf("%s:%s:%s:%s", u.Groupname, "x", u.Gid, "")

}

// Perform a number of sanity checks before beginning
func (u *User) PreFlightChecks() error {

        // Check if a username has been set, this is the only required field
        // f unset throw an error
        if u.Username == "" {
                return errors.New("Username unset")
        } else {
                if _, err := user.Lookup(u.Username); err == nil {
                        return errors.New("User name already in use")
                }
        }

        // If the user id is unset, find the first avail id
        if u.Uid == "" {
                for i := 1; i <= 65536; i++ {
                        if _, err := user.LookupId(strconv.Itoa(i)); err != nil {
                                u.Uid = strconv.Itoa(i)
                                break
                        }
                }
        } else {
                // check if the requested user id is in use
                _, err := user.LookupId(u.Uid)
                if err == nil {
                        return errors.New("User ID already in use")
                }
        }

        if u.Gid == "" {
                for i := 1; i <= 65536; i++ {
                        if _, err := user.LookupGroupId(strconv.Itoa(i)); err != nil {
                                u.Gid = strconv.Itoa(i)
                                break
                        }
                }
        } else {
                if _, err := user.LookupGroupId(u.Gid); err == nil {
                        return errors.New("Group Id already in use")
                }
        }

        // If the group name isnt set, use the username
        if u.Groupname == "" {
                u.Groupname = u.Username
        }

        // Lookup if the group name is in use
        if _, err := user.LookupGroup(u.Groupname); err == nil {
                return errors.New("Group Name already in use")
        }

        // Check if Homedir is set
        if u.Homedir == "" {
                u.Homedir = fmt.Sprintf("/home/%s", u.Username)
        }

        // Check if the users homedir already exists
        if _, err := os.Stat(u.Homedir); !os.IsNotExist(err) {
                return errors.New(fmt.Sprintf("Home directory %s already exists", u.Homedir))
        }

        // Check if a Shell has been set, otherwise
        // default to nologin
        if u.Shell == "" {
                u.Shell = "/sbin/nologin"
        }

        // Check that additional groups exist
        for _, group := range u.GroupMemberships {
                if _, err := user.LookupGroup(group); err != nil {
                        return errors.New(fmt.Sprintf("Additional Group doesnt exist: %s", group))
                }
        }

        return nil

}

// Walk the home diretory of the user setting the corretc ownership
func (u *User) recursiveChownHome() error {
        return filepath.Walk(u.Homedir, func(name string, info os.FileInfo, err error) error {
                if err == nil {
                        uid, _ := strconv.Atoi(u.Uid)
                        gid, _ := strconv.Atoi(u.Gid)
                        err = os.Chown(name, uid, gid)
                }
                return err
        })
}
