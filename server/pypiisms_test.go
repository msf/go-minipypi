package main

import "testing"

func TestNameNormalize(t *testing.T) {
	testNames := func(orig, target string) {
		name := normalizeFileName(orig)
		if name != target {
			t.Errorf("name should be %v, but is: %v", target, name)
		}
	}

	testNames("PyCrypto", "pycrypto")
	testNames("Django-HStore", "django-hstore")
	testNames("django_nose", "django-nose")
	testNames("imposm.parser", "imposm-parser")
}

func TestPackageUrlPathToGetFilePath(t *testing.T) {
	origPath := "/python-dateutil/python_dateutil-2.4.2-py2.py3-none-any.whl"
	wanted := "/python_dateutil-2.4.2-py2.py3-none-any.whl"
	path := handlePypiFileNames(origPath)
	if path != wanted {
		t.Errorf("failed to convert filename, wanted: %v, got: %v", wanted, path)
	}

	origPath = "/imposm-parser/imposm.parser-1.0.6-cp27-none-linux_x86_64.whl"
	wanted = "/imposm.parser-1.0.6-cp27-none-linux_x86_64.whl"
	path = handlePypiFileNames(origPath)
	if path != wanted {
		t.Errorf("failed to convert filename, wanted: %v, got: %v", wanted, path)
	}
}
