package tobs_cli_tests

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/timescale/tobs/cli/pkg/k8s"
)

func testInstall(t testing.TB, name, namespace, filename string, enableBackUp bool) {
	cmds := []string{"install", "--chart-reference", PATH_TO_CHART}
	if name != "" {
		cmds = append(cmds, "-n", name)
	} else {
		cmds = append(cmds, "-n", RELEASE_NAME)
	}
	if namespace != "" {
		cmds = append(cmds, "--namespace", namespace)
	} else {
		cmds = append(cmds, "--namespace", NAMESPACE)
	}
	if filename != "" {
		cmds = append(cmds, "-f", filename)
	}

	if enableBackUp {
		cmds = append(cmds, "--enable-timescaledb-backup")
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	install := exec.Command(PATH_TO_TOBS, cmds...)

	out, err := install.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testHelmInstall(t testing.TB, name, namespace, filename string) {
	cmds := []string{"helm", "install", "--chart-reference", PATH_TO_CHART}
	if name != "" {
		cmds = append(cmds, "-n", name)
	} else {
		cmds = append(cmds, "-n", RELEASE_NAME)
	}
	if namespace != "" {
		cmds = append(cmds, "--namespace", namespace)
	} else {
		cmds = append(cmds, "--namespace", NAMESPACE)
	}
	if filename != "" {
		cmds = append(cmds, "-f", filename)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	install := exec.Command(PATH_TO_TOBS, cmds...)

	out, err := install.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testUninstall(t testing.TB, name, namespace string, deleteData bool) {
	cmds := []string{"uninstall"}
	if name != "" {
		cmds = append(cmds, "-n", name)
	} else {
		cmds = append(cmds, "-n", RELEASE_NAME)
	}
	if namespace != "" {
		cmds = append(cmds, "--namespace", namespace)
	} else {
		cmds = append(cmds, "--namespace", NAMESPACE)
	}
	if deleteData {
		cmds = append(cmds, "--delete-data")
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	uninstall := exec.Command(PATH_TO_TOBS, cmds...)

	out, err := uninstall.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testHelmUninstall(t testing.TB, name, namespace string, deleteData bool) {
	cmds := []string{"helm", "uninstall"}
	if name != "" {
		cmds = append(cmds, "-n", name)
	} else {
		cmds = append(cmds, "-n", RELEASE_NAME)
	}
	if namespace != "" {
		cmds = append(cmds, "--namespace", namespace)
	} else {
		cmds = append(cmds, "--namespace", NAMESPACE)
	}
	if deleteData {
		cmds = append(cmds, "--delete-data")
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	uninstall := exec.Command(PATH_TO_TOBS, cmds...)

	out, err := uninstall.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}

	pods, err := k8s.KubeGetAllPods("tobs", "default")
	if err != nil {
		t.Fatal(err)
	}
	if len(pods) != 0 {
		t.Fatal("Pod remaining after uninstall")
	}

}

func testHelmDeleteData(t testing.TB, name, namespace string) {
	cmds := []string{"helm", "delete-data"}
	if name != "" {
		cmds = append(cmds, "-n", name)
	} else {
		cmds = append(cmds, "-n", RELEASE_NAME)
	}
	if namespace != "" {
		cmds = append(cmds, "--namespace", namespace)
	} else {
		cmds = append(cmds, "--namespace", NAMESPACE)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	deletedata := exec.Command(PATH_TO_TOBS, cmds...)

	out, err := deletedata.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}

	pvcs, err := k8s.KubeGetPVCNames("default", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(pvcs) != 0 {
		t.Fatal("PVC remaining")
	}
}

func testHelmShowValues(t testing.TB) {
	var showvalues *exec.Cmd

	t.Logf("Running 'tobs helm show-values'")

	showvalues = exec.Command(PATH_TO_TOBS, "helm", "show-values", "-c", PATH_TO_CHART)
	out, err := showvalues.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func TestInstallation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping installation tests")
	}

	testHelmShowValues(t)

	testUninstall(t, "", "", true)

	testInstall(t, "abc", "", "", false)
	testHelmUninstall(t, "abc", "", false)

	testHelmInstall(t, "def", "", "")
	testUninstall(t, "def", "", false)
	testHelmDeleteData(t, "def", "")

	testInstall(t, "f1", "", "./../testdata/f1.yml", false)
	testHelmUninstall(t, "f1", "", false)

	testHelmInstall(t, "f2", "", "./../testdata/f2.yml")
	testUninstall(t, "f2", "", false)

	testHelmInstall(t, "f3", "nas", "./../testdata/f3.yml")
	testHelmUninstall(t, "f3", "nas", false)

	testInstall(t, "f4", "", "./../testdata/f4.yml", false)
	testUninstall(t, "f4", "", false)

	testInstall(t, "", "", "", false)

	time.Sleep(1 * time.Minute)

	t.Logf("Waiting for pods to initialize...")
	pods, err := k8s.KubeGetAllPods(NAMESPACE, RELEASE_NAME)
	if err != nil {
		t.Logf("Error getting all pods")
		t.Fatal(err)
	}

	for _, pod := range pods {
		err = k8s.KubeWaitOnPod(NAMESPACE, pod.Name)
		if err != nil {
			t.Logf("Error while waiting on pod")
			t.Fatal(err)
		}
	}

	time.Sleep(30 * time.Second)
}