package vcs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Canary test to ensure HgRepo implements the Repo interface.
var _ Repo = &HgRepo{}

// To verify hg is working we perform integration testing
// with a known hg service.

func TestHg(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "go-vcs-hg-tests")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	repo, err := NewHgRepo("http://hg.code.sf.net/p/vcstesthgrepo/code", tempDir+"/testhgrepo")
	if err != nil {
		t.Error(err)
	}

	if repo.Vcs() != Hg {
		t.Error("Hg is detecting the wrong type")
	}

	// Check the basic getters.
	if repo.Remote() != "http://hg.code.sf.net/p/vcstesthgrepo/code" {
		t.Error("Remote not set properly")
	}
	if repo.LocalPath() != tempDir+"/testhgrepo" {
		t.Error("Local disk location not set properly")
	}

	//Logger = log.New(os.Stdout, "", log.LstdFlags)

	// Do an initial clone.
	err = repo.Get()
	if err != nil {
		t.Errorf("Unable to clone Hg repo. Err was %s", err)
	}

	// Verify Hg repo is a Hg repo
	if !repo.CheckLocal() {
		t.Error("Problem checking out repo or Hg CheckLocal is not working")
	}

	// Test internal lookup mechanism used outside of Hg specific functionality.
	ltype, err := DetectVcsFromFS(tempDir + "/testhgrepo")
	if err != nil {
		t.Error("detectVcsFromFS unable to Hg repo")
	}
	if ltype != Hg {
		t.Errorf("detectVcsFromFS detected %s instead of Hg type", ltype)
	}

	// Test NewRepo on existing checkout. This should simply provide a working
	// instance without error based on looking at the local directory.
	nrepo, nrerr := NewRepo("http://hg.code.sf.net/p/vcstesthgrepo/code", tempDir+"/testhgrepo")
	if nrerr != nil {
		t.Error(nrerr)
	}
	// Verify the right oject is returned. It will check the local repo type.
	if !nrepo.CheckLocal() {
		t.Error("Wrong version returned from NewRepo")
	}

	v, err := repo.Current()
	if err != nil {
		t.Errorf("Error trying Hg Current: %s", err)
	}
	if v != "default" {
		t.Errorf("Current failed to detect Hg on tip of default. Got version: %s", v)
	}

	// Set the version using the short hash.
	err = repo.UpdateVersion("059e82823f3e")
	if err != nil {
		t.Errorf("Unable to update Hg repo version. Err was %s", err)
	}

	// Use Version to verify we are on the right version.
	v, err = repo.Version()
	if v != "059e82823f3ec23fceb532621ee7c15d5624f35f" {
		t.Errorf("Error checking checked out Hg version: %s", v)
	}
	if err != nil {
		t.Error(err)
	}

	v, err = repo.Current()
	if err != nil {
		t.Errorf("Error trying Hg Current for ref: %s", err)
	}
	if v != "059e82823f3ec23fceb532621ee7c15d5624f35f" {
		t.Errorf("Current failed to detect Hg on ref of branch. Got version: %s", v)
	}

	// Use Date to verify we are on the right commit.
	d, err := repo.Date()
	if err != nil {
		t.Error(err)
	}
	if d.Format(longForm) != "2022-03-21 15:53:47 -0400" {
		t.Error("Error checking checked out Hg commit date. Got wrong date:", d)
	}

	// Perform an update.
	err = repo.Update()
	if err != nil {
		t.Error(err)
	}

	v, err = repo.Version()
	if v != "c068171728d3cc343fef21ffdff43cdb14e3c716" {
		t.Errorf("Error checking checked out Hg version: %s", v)
	}
	if err != nil {
		t.Error(err)
	}

	tags, err := repo.Tags()
	if err != nil {
		t.Error(err)
	}
	if tags[1] != "1.0.0" {
		t.Error("Hg tags is not reporting the correct version")
	}

	tags, err = repo.TagsFromCommit("059e82823f3e")
	if err != nil {
		t.Error(err)
	}
	if len(tags) != 0 {
		t.Error("Hg is incorrectly returning tags for a commit")
	}

	tags, err = repo.TagsFromCommit("96379533a643")
	if err != nil {
		t.Error(err)
	}
	if len(tags) != 1 || tags[0] != "1.0.0" {
		t.Error("Hg is incorrectly returning tags for a commit")
	}

	branches, err := repo.Branches()
	if err != nil {
		t.Error(err)
	}
	// The branches should be HEAD, master, and test.
	if branches[0] != "test" {
		t.Error("Hg is incorrectly returning branches")
	}

	if !repo.IsReference("1.0.0") {
		t.Error("Hg is reporting a reference is not one")
	}

	if !repo.IsReference("test") {
		t.Error("Hg is reporting a reference is not one")
	}

	if repo.IsReference("foo") {
		t.Error("Hg is reporting a non-existent reference is one")
	}

	if repo.IsDirty() {
		t.Error("Hg incorrectly reporting dirty")
	}

	ci, err := repo.CommitInfo("72a363187366")
	if err != nil {
		t.Error(err)
	}
	if ci.Commit != "72a3631873669f4bd4c41e4d9146104e1e55e767" {
		t.Error("Hg.CommitInfo wrong commit id")
	}
	if ci.Author != "Matt Farina <matt@mattfarina.com>" {
		t.Error("Hg.CommitInfo wrong author")
	}
	if ci.Message != "Removing a" {
		t.Error("Hg.CommitInfo wrong message")
	}

	ti := time.Unix(1647898961, 0)
	if !ti.Equal(ci.Date) {
		t.Error("Hg.CommitInfo wrong date")
	}

	_, err = repo.CommitInfo("asdfasdfasdf")
	if err != ErrRevisionUnavailable {
		t.Error("Hg didn't return expected ErrRevisionUnavailable")
	}

	tempDir2, err := ioutil.TempDir("", "go-vcs-hg-tests-export")
	if err != nil {
		t.Fatalf("Error creating temp directory: %s", err)
	}
	defer func() {
		err = os.RemoveAll(tempDir2)
		if err != nil {
			t.Error(err)
		}
	}()

	exportDir := filepath.Join(tempDir2, "src")

	err = repo.ExportDir(exportDir)
	if err != nil {
		t.Errorf("Unable to export Hg repo. Err was %s", err)
	}

	_, err = os.Stat(filepath.Join(exportDir, "README.md"))
	if err != nil {
		t.Errorf("Error checking exported file in Hg: %s", err)
	}

	_, err = os.Stat(filepath.Join(exportDir, string(repo.Vcs())))
	if err != nil {
		if found := os.IsNotExist(err); !found {
			t.Errorf("Error checking exported metadata in Hg: %s", err)
		}
	} else {
		t.Error("Error checking Hg metadata. It exists.")
	}
}

func TestHgCheckLocal(t *testing.T) {
	// Verify repo.CheckLocal fails for non-Hg directories.
	// TestHg is already checking on a valid repo
	tempDir, err := ioutil.TempDir("", "go-vcs-hg-tests")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	repo, _ := NewHgRepo("", tempDir)
	if repo.CheckLocal() {
		t.Error("Hg CheckLocal does not identify non-Hg location")
	}

	// Test NewRepo when there's no local. This should simply provide a working
	// instance without error based on looking at the remote localtion.
	_, nrerr := NewRepo("http://hg.code.sf.net/p/vcstesthgrepo/code", tempDir+"/testhgrepo")
	if nrerr != nil {
		t.Error(nrerr)
	}
}

func TestHgPing(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "go-vcs-hg-tests")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	repo, err := NewHgRepo("http://hg.code.sf.net/p/vcstesthgrepo/code", tempDir)
	if err != nil {
		t.Error(err)
	}

	ping := repo.Ping()
	if !ping {
		t.Error("Hg unable to ping working repo")
	}

	repo, err = NewHgRepo("http://bitbucket.org/mattfarina/ihopethisneverexistsbecauseitshouldnt", tempDir)
	if err != nil {
		t.Error(err)
	}

	ping = repo.Ping()
	if ping {
		t.Error("Hg got a ping response from when it should not have")
	}
}

func TestHgInit(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "go-vcs-hg-tests")
	repoDir := tempDir + "/repo"
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	repo, err := NewHgRepo(repoDir, repoDir)
	if err != nil {
		t.Error(err)
	}

	err = repo.Init()
	if err != nil {
		t.Error(err)
	}

	v, err := repo.Version()
	if err != nil {
		t.Error(err)
	}
	if !strings.HasPrefix(v, "000000") {
		t.Errorf("Hg Init reporting wrong initial version: %s", v)
	}
}
