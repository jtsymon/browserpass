package pass

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func TestDefaultStorePath(t *testing.T) {
	var home, expected, actual string

	usr, err := user.Current()

	if err != nil {
		t.Error(err)
	}

	home = usr.HomeDir

	// default directory
	os.Setenv("PASSWORD_STORE_DIR", "")
	expected = filepath.Join(home, ".password-store")
	actual, _ = defaultStorePath()
	if expected != actual {
		t.Errorf("%s does not match %s", expected, actual)
	}

	// custom directory from $PASSWORD_STORE_DIR
	expected, err = filepath.Abs("browserpass-test")
	if err != nil {
		t.Error(err)
	}

	fmt.Println(expected)
	os.Mkdir(expected, os.ModePerm)
	os.Setenv("PASSWORD_STORE_DIR", expected)
	actual, _ = defaultStorePath()
	if expected != actual {
		t.Errorf("%s does not match %s", expected, actual)
	}

	// clean-up
	os.Setenv("PASSWORD_STORE_DIR", "")
	os.Remove(expected)
}

func TestDiskStore_Search_nomatch(t *testing.T) {
	s, err := NewDefaultStore()
	if err != nil {
		t.Fatal(err)
	}

	domain := "this-most-definitely-does-not-exist"
	logins, err := s.Search(domain)
	if err != nil {
		t.Fatal(err)
	}
	if len(logins) > 0 {
		t.Errorf("%s yielded results, but it should not", domain)
	}
}

func TestDiskStoreSearch(t *testing.T) {
	store := diskStore{"test_store"}
	targetDomain := "abc.com"
	testDomains := []string{"abc.com", "test.abc.com", "testing.test.abc.com"}
	for _, domain := range testDomains {
		searchResults, err := store.Search(domain)
		if err != nil {
			t.Fatal(err)
		}
		// check if result contains abc.com
		found := false
		for _, searchResult := range searchResults {
			if searchResult == targetDomain {
				found = true
				break
			}
		}
		if found != true {
			t.Fatalf("Couldn't find %v in %v", targetDomain, searchResults)
		}
	}
}

func TestDiskStoreSearchNoDuplicatesWhenPatternMatchesDomainAndUsername(t *testing.T) {
	store := diskStore{"test_store"}
	searchResult, err := store.Search("xyz")
	if err != nil {
		t.Fatal(err)
	}
	if len(searchResult) != 1 {
		t.Fatalf("Found %v results instead of 1", len(searchResult))
	}
	expectedResult := "xyz.com/xyz_user"
	if searchResult[0] != expectedResult {
		t.Fatalf("Couldn't find %v, found %v instead", expectedResult, searchResult[0])
	}
}

func TestDiskStoreSearchFollowsSymlinkFiles(t *testing.T) {
	store := diskStore{"test_store"}
	searchResult, err := store.Search("def.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(searchResult) != 1 {
		t.Fatalf("Found %v results instead of 1", len(searchResult))
	}
	expectedResult := "def.com"
	if searchResult[0] != expectedResult {
		t.Fatalf("Couldn't find %v, found %v instead", expectedResult, searchResult[0])
	}
}

func TestDiskStoreSearchFollowsSymlinkDirectories(t *testing.T) {
	store := diskStore{"test_store"}
	searchResult, err := store.Search("amazon.co.uk")
	if err != nil {
		t.Fatal(err)
	}
	if len(searchResult) != 2 {
		t.Fatalf("Found %v results instead of 2", len(searchResult))
	}
	expectedResult := []string{"amazon.co.uk/user1", "amazon.co.uk/user2"}
	if searchResult[0] != expectedResult[0] || searchResult[1] != expectedResult[1] {
		t.Fatalf("Couldn't find %v, found %v instead", expectedResult, searchResult)
	}
}
