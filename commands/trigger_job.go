package commands

import (
	"fmt"
	"log"

	"github.com/concourse/fly/commands/internal/flaghelpers"
	"github.com/concourse/fly/rc"
	"github.com/concourse/go-concourse/concourse"
)

type TriggerJobCommand struct {
	Job flaghelpers.JobFlag `short:"j" long:"job" required:"true"  value-name:"PIPELINE/JOB"   description:"Name of a job to hijack"`
}

func (command *TriggerJobCommand) Execute([]string) error {
	connection, err := rc.TargetConnection(Fly.Target)
	if err != nil {
		log.Fatalln(err)
	}
	client := concourse.NewClient(connection)

	build, err := client.CreateJobBuild(command.Job.PipelineName, command.Job.JobName)
	if err != nil {
		log.Fatalln("Something unexpected happened with the server")
	}

	fmt.Printf("build %s #%s has started\n", build.JobName, build.Name)
	fmt.Println("To watch your build run, run the following command:")
	fmt.Printf("fly -t %s watch --job %s/%s --build %s\n",
		Fly.Target,
		command.Job.PipelineName,
		command.Job.JobName,
		build.Name,
	)
	return nil
}
