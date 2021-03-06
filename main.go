package main 


import (
    "fmt"
    "os"
    "os/user"
    "io/ioutil"
    "strings"
)

type Config struct {
    TaskbookRoot        string      // where all the taskbooks are stored in, $HOME/.taskbook
    TBMProfilePath      string      // path to tbm.profile 
    TaskbookProfile     string      // the currently active profile, gets assigned when tbm.profile is read, default: undefined
    StoragePath         string      // .taskbook/storage/
    StorageActive       bool        // .taskbook/storage/storage.json -> true if exists
    ArchivePath         string      // .taskbook/archive/
    ArchiveActive       bool        // .taskbook/archive/archive.json -> true if exists
    Method              string      // method to call (eg. switch)
    Argument            string      // Passed argument
}


// Execute the program, make sure everything is setup correctly.
func (c *Config) run() error {
    err := pathExists(c.TaskbookRoot)
    if err != nil {
        return err
    } 

    err = loadProfile(c)
    if err != nil {
        return err
    }
 
    // Check if the storage and archive exists.
    err = pathExists(c.StoragePath)
    if err != nil {
        return err
    }

    // Check if the storage.json file exists.
    err = pathExists(c.StoragePath + "/storage.json")
    if err != nil {
        return err
    } else {
        c.StorageActive = true
    }

    // Archives aren't that important.
    err = pathExists(c.ArchivePath)
    if err != nil {
        fmt.Printf("Archive does not exist, continuing.\n")
    } else {
        err = pathExists(c.ArchivePath + "/archive.json")
        if err != nil {
            fmt.Printf("Archive is not active, continuing.\n")
        } else {
            c.ArchiveActive = true
        }
    }

    // Execute the given method.
    switch c.Method {
    case "switch":
        err = switchProfile(c.StoragePath, c.ArchivePath, c.TaskbookProfile, c.TBMProfilePath, c.Argument)
        if err != nil {
            return err
        }
    case "rename":
        err = writeProfile(c.TBMProfilePath, c.Argument)
        if err != nil {
            return err
        }
    case "new":
        var newProfileName string
        if c.Argument == " " || c.Argument == "-" || c.Argument == "storage" || c.Argument == "archive" {
            newProfileName = ""
        } else {
            newProfileName = c.Argument
        }
        err = newProfile(c.StoragePath, c.ArchivePath, newProfileName)
        if err != nil {
            return err
        }
    case "profile":
        err = loadProfile(c)
        if err != nil {
            return err
        }
        // return the profile name.
        fmt.Printf("Current profile: %s", c.TaskbookProfile)
    case "all-profiles":
        files, err := ioutil.ReadDir(c.StoragePath)
        if err != nil {
            return err
        }

        fmt.Println("Available profiles:\n")
        for _, f := range files {
            fmt.Printf("- %s\n", f.Name())
        }
    case "help":
        printHelp()
    default:
        return fmt.Errorf("Defined method (%s) is not does not exist, how did you do that?\n", c.Method)
    }
    return nil
}


// Create a new profile.
func newProfile(storage, archive, profileName string) error {
    // Check if a profileName is provided or not.
    // When no name is provided, the new profile will be new<id>.
    // <id> will be a sequential number.
    // eg. if new exists, new1 will be created and so on.
    if profileName == "" {
        var id int
        for {
            storErr := pathExists(storage + "/new" + string(id) + ".json")
            archErr := pathExists(archive + "/new" + string(id) + ".json")
            if storErr == nil && archErr == nil {
                profileName = "new" + string(id) + ".json"
                break
            }
            id++
        }
    } else {
        profileName = profileName + ".json"
    }

    // Write the profiles
    err := writeProfile(storage + "/" + profileName, "{}")   
    if err != nil {
        return err
    }
    err = writeProfile(archive + "/" + profileName, "{}")   
    if err != nil {
        return err
    }
    return nil
}


// Switch to another profile.
func switchProfile(storage, archive, profile, profilePath, arg string) error {
    // Check if the "switchTo" profile even exists.
    argJson := arg + ".json"
    argArchive := true

    // Make sure the <arg>.json files exist.
    err := pathExists(storage + "/" + argJson)
    if err != nil {
        return err
    }
    err = pathExists(archive + "/" + argJson)
    if err != nil {
        fmt.Printf("Archive (%s) does not exist, continuing.\n", archive + "/" + argJson)
        argArchive = false
    }

    // Make sure the storage.json & archive.json exists.
    // If so, rename the taskbooks
    err = pathExists(storage + "/storage.json")
    if err != nil {
        return err
    } else {
        err = os.Rename(storage + "/storage.json", storage + "/" + profile + ".json")
        if err != nil {
            return fmt.Errorf("An error occured while changing storage.json, error: %s\n", err)
        }
    }
    err = pathExists(archive + "/archive.json")
    if err != nil {
        fmt.Printf("Archive (%s) doesn't exist, skip.\n", archive + "/archive.json")
    } else {
        err = os.Rename(archive + "/archive.json", archive + "/" + profile + ".json")
        if err != nil {
            return fmt.Errorf("An error occured while changing archive.json, error: %s\n", err)
        }
    }

    // Rename the profile you want to switch to the required names (eg. dev.json -> storage.json)
    err = os.Rename(storage + "/" + argJson, storage + "/storage.json")
    if err != nil {
        return fmt.Errorf("An error occured while changing %s to storage.json, error: %s\n", argJson, err)
    }
    if argArchive {
        err = os.Rename(archive + "/" + argJson, archive + "/archive.json")
        if err != nil {
            return fmt.Errorf("An error occured while changing %s to storage.json, error: %s\n", argJson, err)
        }
    }
    // Update the TBMProfile
    err = writeProfile(profilePath, arg)
    if err != nil {
        return err
    }

    return nil
}


// Creates/Writes profile (string) to file.
func writeProfile(path, data string) error {
    file, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("Could not create %s, error: %s\n", path, err)
    }
    defer file.Close()

    // write data to file (path)
    _, err = file.WriteString(data)
    if err != nil {
        return fmt.Errorf("Something went wrong while writing (%s) to file (%s), error: %s\n", data, path, err)
    }
    return nil
}


// Load the currently *active* profile.
func loadProfile(c *Config) error {
    var TBMProfileExists = true 

    err := pathExists(c.TBMProfilePath)
    if err != nil {
        TBMProfileExists = false
    }

    if TBMProfileExists {
        data, err := ioutil.ReadFile(c.TBMProfilePath)
        if err != nil {
            return fmt.Errorf("Could not get data from tbm.profile (%s), error: %s\n", c.TBMProfilePath, err)
        }
        if string(data) == "" {
            fmt.Printf("Taskbook profile name cannot be empty, creating default profile.\n")
            err = writeProfile(c.TBMProfilePath, "default")
            if err != nil {
                return fmt.Errorf("Could not create tbm.profile (%s), error: %s\n", c.TBMProfilePath, err)
            }
        } 
        c.TaskbookProfile = string(data)
    } else {
        err = writeProfile(c.TBMProfilePath, "default")
        if err != nil {
            return fmt.Errorf("Could not create tbm.profile (%s), error: %s\n", c.TBMProfilePath, err)
        }
    }

    return nil 
}


// Check if the given path exists, else, return error.
func pathExists(path string) error {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return fmt.Errorf("Path (%s) does not exist, error: %s\n", path, err)
    }
    return nil
}


// Build/construct the configuration.
func constructor() (Config, error) {
    var method string

    // Get information about the current user.
    user, err := user.Current()
    if err != nil {
        return Config{}, fmt.Errorf("Could not get current user information, error: %s\n", err)
    }
    // Get arguments without the program name [1:]
    cmd_args := os.Args[1:]
    if len(cmd_args) <= 0 {
       return Config{}, fmt.Errorf("No paramaters specified!\n\nTry 'tbm help' for more information.\n") 
    }

    // Make sure a valid method is provided.
    // This can be extended further on, define a method here and in the run() function.
    switch strings.ToLower(cmd_args[0]) {
    case "switch", "-s", "--switch":
        method = "switch"
    case "rename", "-r", "--rename":
        method = "rename"
    case "new", "-n", "--new":
        method = "new"
    case "profile", "-p", "--profile":
        method = "profile"
        cmd_args = append(cmd_args, "-")
    case "all-profiles", "-a", "--all-profiles":
        method = "all-profiles"
        cmd_args = append(cmd_args, "-")
    case "help", "-h", "--help":
        method = "help"
        cmd_args = append(cmd_args, "-") 
    default:
        // if no method was provided, exit the program.
        return Config{}, fmt.Errorf("Method not known\n\nTry 'tbm help' for more information.\n")
    }
    dotTaskbook := user.HomeDir + "/.taskbook"

    // Define basic information for the system.
    config := Config{
        dotTaskbook,                    // TaskbookRoot
        dotTaskbook + "/tbm.profile",   // TBMProfilePath
        "undefined",                    // TaskbookProfile
        dotTaskbook + "/storage",       // Storage
        false,                          // StorageActive
        dotTaskbook + "/archive",       // Archive
        false,                          // StorageActive
        method,                         // Method
        cmd_args[1],                    // Argument
    }
    return config, nil 
}

func printHelp() {
    // the format looks like shit but hey, it works ¯\_(ツ)_/¯
        fmt.Printf(`
tbm is a little helper tool that extends the functionality of taskbook.
Plese note, you always need to provide a method but not always args.
If something is optional, it's surrounded with '< >'. 

Usage: tbm method <args>

Method(s):
 - switch:          Switch/Change to another profile.
                     > tbm -s private / tbm switch private
 - rename:          Rename the current profile.
                     > tbm -r default / tbm rename default
 - new:             Create a new profile.
                     > tbm -n <new-defaul> / tbm new <new-default>
 - profile:         Shows the current profile.
                     > tbm -p / tbm profile
 - all-profiles:    Get ALL the available profiles.
                     > tbm -a / tbm all-profiles
 - help:            This message.
                     > tbm -h / tbm help
`)
}

func main() {
    // Construct the configuration.
    config, err := constructor()
    if err != nil {
        fmt.Printf("An error occured while constructing\nError: %s\n", err)
        os.Exit(1)
    }
    err = config.run()
    if err != nil {
        fmt.Printf("Something went wrong while preparing\nError: %s\n", err)
        os.Exit(1)
    }
    // if the DEBUG file exists, make a config dump
    err = pathExists("./DEBUG")
    if err == nil {
        // DEBUG: Make a config-dump so i can see what is assigned.
        fmt.Printf("\nConfig-Dump:\n- TaskbookRoot: %s\n- TBMProfilePath: %s\n- TaskbookProfile: %s\n- Storage: %s\n- StorageActive: %t\n- Archive: %s\n- ArchiveActive: %t\n- Method: %s\n- Argument: %s\n", config.TaskbookRoot, config.TBMProfilePath, config.TaskbookProfile, config.StoragePath, config.StorageActive, config.ArchivePath, config.ArchiveActive, config.Method, config.Argument)
    }

    os.Exit(0)
}
