package mainpackage main



import (import (

	"encoding/json""encoding/json"

	"fmt""fmt"

	"flyby/internal/concourse""flyby/internal/concourse"

	"os/exec""os/exec"

))



func main() {func main() {

	cmd := exec.Command("fly", "-t", "concourse-prod-fin-ca", "pipelines", "--json")cmd := exec.Command("fly", "-t", "concourse-prod-fin-ca", "pipelines", "--json")

	output, err := cmd.Output()output, err := cmd.Output()

	if err != nil {if err != nil {

		fmt.Printf("Command error: %v\n", err)fmt.Printf("Command error: %v

		return", err)

	}return

	}

	var pipelines []concourse.Pipeline

	err = json.Unmarshal(output, &pipelines)var pipelines []concourse.Pipeline

	if err != nil {err = json.Unmarshal(output, &pipelines)

		fmt.Printf("JSON error: %v\n", err)if err != nil {

		returnfmt.Printf("JSON error: %v

	}", err)

	return

	fmt.Printf("Successfully parsed %d pipelines\n", len(pipelines))}

	if len(pipelines) > 0 {

		p := pipelines[0]fmt.Printf("Successfully parsed %d pipelines

		fmt.Printf("First pipeline: %s (last updated: %s)\n", p.Name, p.GetLastUpdated().Format("2006-01-02 15:04:05"))", len(pipelines))

	}if len(pipelines) > 0 {

}p := pipelines[0]
fmt.Printf("First pipeline: %s (last updated: %s)
", p.Name, p.GetLastUpdated().Format("2006-01-02 15:04:05"))
}
}
