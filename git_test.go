package vcs

import (
	"io/ioutil"
	//"log"
	"os"
	"testing"
)

// Canary test to ensure GitRepo implements the Repo interface.
var _ Repo = &GitRepo{}

// To verify git is working we perform intergration testing
// with a known git service.

func TestGit(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "go-vcs-git-tests")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	repo := NewGitRepo("https://github.com/Masterminds/VCSTestRepo", tempDir+"/VCSTestRepo")

	// Check the basic getters.
	if repo.Remote() != "https://github.com/Masterminds/VCSTestRepo" {
		t.Error("Remote not set properly")
	}
	if repo.LocalPath() != tempDir+"/VCSTestRepo" {
		t.Error("Local disk location not set properly")
	}

	//Logger = log.New(os.Stdout, "", log.LstdFlags)

	// Do an initial clone.
	err = repo.Get()
	if err != nil {
		t.Errorf("Unable to clone Git repo. Err was %s", err)
	}

	// Verify Git repo is a Git repo
	if repo.CheckLocal() == false {
		t.Error("Problem checking out repo or Git CheckLocal is not working")
	}

	// Perform an update.
	err = repo.Update()
	if err != nil {
		t.Error(err)
	}

	// Set the version using the short hash.
	err = repo.UpdateVersion("806b07b")
	if err != nil {
		t.Errorf("Unable to update Git repo version. Err was %s", err)
	}

	// Use Version to verify we are on the right version.
	v, err := repo.Version()
	if v != "806b07b08faa21cfbdae93027904f80174679402" {
		t.Error("Error checking checked out Git version")
	}
	if err != nil {
		t.Error(err)
	}

	// Verify that we can set the version something other than short hash
	err = repo.UpdateVersion("master")
	if err != nil {
		t.Errorf("Unable to update Git repo version. Err was %s", err)
	}
	err = repo.UpdateVersion("806b07b08faa21cfbdae93027904f80174679402")
	if err != nil {
		t.Errorf("Unable to update Git repo version. Err was %s", err)
	}
	v, err = repo.Version()
	if v != "806b07b08faa21cfbdae93027904f80174679402" {
		t.Error("Error checking checked out Git version")
	}
	if err != nil {
		t.Error(err)
	}

}

func TestGitCheckLocal(t *testing.T) {
	// Verify repo.CheckLocal fails for non-Git directories.
	// TestGit is already checking on a valid repo
	tempDir, err := ioutil.TempDir("", "go-vcs-git-tests")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	repo := NewGitRepo("", tempDir)
	if repo.CheckLocal() == true {
		t.Error("Git CheckLocal does not identify non-Git location")
	}
}