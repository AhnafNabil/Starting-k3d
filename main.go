package main

import (
	"log"
	"os"

	run "github.com/AhnafNabil/Starting-k3d/cli"
	"github.com/AhnafNabil/Starting-k3d/version"
	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "k3d"
	app.Usage = "Run k3s in Docker!"
	app.Version = version.GetVersion()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "iwilltry42",
			Email: "iwilltry42@gmail.com",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "check-tools",
			Aliases: []string{"ct"},
			Usage:   "Check if docker is running",
			Action:  run.CheckTools,
		},
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "Create a single node k3s cluster in a container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: "k3s_default",
					Usage: "Set a name for the cluster",
				},
				cli.StringFlag{
					Name:  "volume, v",
					Usage: "Mount a volume into the cluster node (Docker notation: `source:destination`)",
				},
				cli.StringFlag{
					Name:  "version",
					Value: "v1.29.4-rc1-k3s1",
					Usage: "Choose the k3s image version",
				},
				cli.IntFlag{
					Name:  "port, p",
					Value: 6443,
					Usage: "Set a port on which the ApiServer will listen",
				},
				cli.IntFlag{
					Name:  "timeout, t",
					Value: 0,
					Usage: "Set the timeout value when --wait flag is set",
				},
				cli.BoolFlag{
					Name:  "wait, w",
					Usage: "Wait for the cluster to come up",
				},
			},
			Action: run.CreateCluster,
		},
		{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "Delete cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: "k3s_default",
					Usage: "name of the cluster",
				},
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "delete all existing clusters (this ignores the --name/-n flag)",
				},
			},
			Action: run.DeleteCluster,
		},
		{
			Name:  "stop",
			Usage: "Stop cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: "k3s_default",
					Usage: "name of the cluster",
				},
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "stop all running clusters (this ignores the --name/-n flag)",
				},
			},
			Action: run.StopCluster,
		},
		{
			Name:  "start",
			Usage: "Start a stopped cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: "k3s_default",
					Usage: "name of the cluster",
				},
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "start all stopped clusters (this ignores the --name/-n flag)",
				},
			},
			Action: run.StartCluster,
		},
		{
			Name:    "list",
			Aliases: []string{"ls", "l"},
			Usage:   "List all clusters",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "also show non-running clusters",
				},
			},
			Action: run.ListClusters,
		},
		{
			Name:  "get-kubeconfig",
			Usage: "Get kubeconfig location for cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: "k3s_default",
					Usage: "name of the cluster",
				},
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "get kubeconfig for all clusters (this ignores the --name/-n flag)",
				},
			},
			Action: run.GetKubeConfig,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
