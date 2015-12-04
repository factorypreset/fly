package integration_test

import (
	"fmt"
	"os/exec"

	"github.com/concourse/atc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Fly CLI", func() {
	Describe("trigger-job", func() {
		var (
			flyCmd    *exec.Cmd
			atcServer *ghttp.Server
		)

		BeforeEach(func() {
			atcServer = ghttp.NewServer()
			flyCmd = exec.Command(flyPath, "-t", atcServer.URL(), "trigger-job")
		})

		Context("when it is given a pipeline name and a job name", func() {
			var build atc.Build

			BeforeEach(func() {
				flyCmd.Args = append(flyCmd.Args, "--job", "main/jettison")

				build = atc.Build{
					JobName:      "jettison",
					PipelineName: "main",
					Name:         "1411",
				}

			})

			Context("when the server responds with success", func() {
				BeforeEach(func() {
					atcServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v1/pipelines/main/jobs/jettison/builds"),
							ghttp.RespondWithJSONEncoded(200, build),
						),
					)
				})

				It("build "+build.Name+" has started", func() {
					sess, err := gexec.Start(flyCmd, nil, nil)
					Expect(err).ToNot(HaveOccurred())
					Eventually(sess.Out).Should(gbytes.Say("build jettison #" + build.Name + " has started"))
					Eventually(sess).Should(gexec.Exit(0))
				})

				It("tells the user how to watch the build", func() {
					sess, err := gexec.Start(flyCmd, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					command := fmt.Sprintf("fly -t %s watch --job main/jettison --build 1411", atcServer.URL())

					Eventually(sess.Out).Should(gbytes.Say("To watch your build run, run the following command:"))
					Eventually(sess.Out).Should(gbytes.Say(command))
					Eventually(sess).Should(gexec.Exit(0))
				})
			})

			Context("when the server responds that another unexpected error happened", func() {
				BeforeEach(func() {
					atcServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v1/pipelines/main/jobs/jettison/builds"),
							ghttp.RespondWithJSONEncoded(500, build),
						),
					)
				})

				It("Something unexpected happened with the server", func() {
					sess, err := gexec.Start(flyCmd, nil, nil)
					Expect(err).ToNot(HaveOccurred())
					Eventually(sess.Err).Should(gbytes.Say("Something unexpected happened with the server"))
					Eventually(sess).Should(gexec.Exit(1))
				})
			})
		})

		Context("when the --job flag is missing", func() {
			It("shows an error message", func() {
				sess, err := gexec.Start(flyCmd, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				Eventually(sess.Err).Should(gbytes.Say("error: the required flag `-j, --job' was not specified"))
				Eventually(sess).Should(gexec.Exit(1))
			})
		})

		Context("when pipeline or job name is missing", func() {
			BeforeEach(func() {
				flyCmd.Args = append(flyCmd.Args, "--job")
			})

			It("shows them an error message", func() {
				sess, err := gexec.Start(flyCmd, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				Eventually(sess.Err).Should(gbytes.Say("error: expected argument for flag `-j, --job'"))
				Eventually(sess).Should(gexec.Exit(1))
			})
		})
	})
})
