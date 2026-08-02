package cmd
import(
	"path/filepath"
	"github.com/joho/godotenv"
	"fmt"
	"os"
	
)

func loadEnv() {
	exePath, _ := os.Executable()
	resPath, _ := filepath.EvalSymlinks(exePath)
	binDir := filepath.Dir(resPath)
	envPath := filepath.Join(filepath.Dir(binDir), ".env")
	if err := godotenv.Load(envPath); err != nil {
		fmt.Printf("warning: .env not loaded from %s: %v\n", envPath, err)
	}
}
