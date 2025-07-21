package pr

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "time"

    corev1 "k8s.io/api/core/v1"
)

func CreateOptimizationPR(pod corev1.Pod, suggestions map[string]int64) error {
    repo := "markojeremic/ai-k8s-optimizer"
    branch := fmt.Sprintf("mem-opt-%s-%d", pod.Name, time.Now().Unix())
    dir := "/tmp/mem-optimizer"

    os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    cmds := [][]string{
        {"gh", "repo", "clone", repo, dir},
        {"git", "-C", dir, "checkout", "-b", branch},
    }

    for _, args := range cmds {
        if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
            return fmt.Errorf("error: %v, output: %s", err, out)
        }
    }

    filename := filepath.Join(dir, "suggestions", fmt.Sprintf("%s.yaml", pod.Name))
    f, _ := os.Create(filename)
    defer f.Close()

    for cname, val := range suggestions {
        fmt.Fprintf(f, "%s: %d\n", cname, val)
    }

    cmds = [][]string{
        {"git", "-C", dir, "add", "."},
        {"git", "-C", dir, "commit", "-m", "Suggest memory optimizations"},
        {"git", "-C", dir, "push", "--set-upstream", "origin", branch},
        {"gh", "pr", "create", "--title", "Memory Optimization", "--body", "Suggested memory usage update", "--head", branch, "--base", "main"},
    }

    for _, args := range cmds {
        if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
            return fmt.Errorf("error: %v, output: %s", err, out)
        }
    }

    return nil
}
