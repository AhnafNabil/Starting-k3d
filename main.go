package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/urfave/cli"
)

func createCluster(c *cli.Context) error {
	createClusterDir(c.String("name"))
	port := fmt.Sprintf("%s:%s", c.String("port"), c.String("port"))
	image := fmt.Sprintf("rancher/k3s:%s", c.String("version"))
	cmd := "docker"
	args := []string{
		"run",
		"--name", c.String("name"),
		"-e", "K3S_KUBECONFIG_OUTPUT=/output/kubeconfig.yaml",
		"--publish", port,
		"--privileged",
	}
	extraArgs := []string{}
	if c.IsSet("volume") {
		extraArgs = append(extraArgs, "--volume", c.String("volume"))
	}
	if len(extraArgs) > 0 {
		args = append(args, extraArgs...)
	}
	args = append(args,
		"-d",
		image,
		"server",                                // cmd
		"--https-listen-port", c.String("port"), //args
	)
	log.Printf("Creating cluster [%s]", c.String("name"))
	if err := run(true, cmd, args...); err != nil {
		log.Fatalf("FAILURE: couldn't create cluster [%s] -> %+v", c.String("name"), err)
		return err
	}
	log.Printf("SUCCESS: created cluster [%s]", c.String("name"))
	log.Printf(`You can now use the cluster with:

export KUBECONFIG="$(%s get-kubeconfig --name='%s')"
kubectl cluster-info`, os.Args[0], c.String("name"))
	return nil
}

func deleteCluster(c *cli.Context) error {
	cmd := "docker"
	args := []string{"rm"}
	clusters := []string{}

	// operate on one or all clusters
	if !c.Bool("all") {
		clusters = append(clusters, c.String("name"))
	} else {
		clusterList, err := getClusterNames()
		if err != nil {
			log.Fatalf("ERROR: `--all` specified, but no clusters were found.")
		}
		clusters = append(clusters, clusterList...)
	}

	// remove clusters one by one instead of appending all names to the docker command
	// this allows for more granular error handling and logging
	for _, cluster := range clusters {
		log.Printf("Removing cluster [%s]", cluster)
		args = append(args, cluster)
		if err := run(true, cmd, args...); err != nil {
			log.Printf("WARNING: couldn't delete cluster [%s], trying a force remove now.", cluster)
			args = args[:len(args)-1] // pop last element from list (name of cluster)
			args = append(args, "-f", cluster)
			if err := run(true, cmd, args...); err != nil {
				log.Printf("FAILURE: couldn't delete cluster [%s] -> %+v", cluster, err)
			}
			args = args[:len(args)-1] // pop last element from list (-f flag)
		}
		deleteClusterDir(cluster)
		log.Printf("SUCCESS: removed cluster [%s]", cluster)
		args = args[:len(args)-1] // pop last element from list (name of last cluster)
	}

	return nil
}

func stopCluster(c *cli.Context) error {
	cmd := "docker"
	args := []string{"stop"}
	clusters := []string{}

	// operate on one or all clusters
	if !c.Bool("all") {
		clusters = append(clusters, c.String("name"))
	} else {
		clusterList, err := getClusterNames()
		if err != nil {
			log.Fatalf("ERROR: `--all` specified, but no clusters were found.")
		}
		clusters = append(clusters, clusterList...)
	}

	// stop clusters one by one instead of appending all names to the docker command
	// this allows for more granular error handling and logging
	for _, cluster := range clusters {
		log.Printf("Starting cluster [%s]", cluster)
		args = append(args, cluster)
		if err := run(true, cmd, args...); err != nil {
			log.Printf("FAILURE: couldn't stop cluster [%s] -> %+v", cluster, err)
		}
		log.Printf("SUCCESS: stopped cluster [%s]", cluster)
		args = args[:len(args)-1] // pop last element from list (name of last cluster)
	}
	return nil
}

func startCluster(c *cli.Context) error {
	cmd := "docker"
	args := []string{"start"}
	clusters := []string{}

	// operate on one or all clusters
	if !c.Bool("all") {
		clusters = append(clusters, c.String("name"))
	} else {
		clusterList, err := getClusterNames()
		if err != nil {
			log.Fatalf("ERROR: `--all` specified, but no clusters were found.")
		}
		clusters = append(clusters, clusterList...)
	}

	// start clusters one by one instead of appending all names to the docker command
	// this allows for more granular error handling and logging
	for _, cluster := range clusters {
		log.Printf("Starting cluster [%s]", cluster)
		args = append(args, cluster)
		if err := run(true, cmd, args...); err != nil {
			log.Printf("FAILURE: couldn't start cluster [%s] -> %+v", cluster, err)
		}
		log.Printf("SUCCESS: started cluster [%s]", cluster)
		args = args[:len(args)-1] // pop last element from list (name of last cluster)
	}
	return nil
}

func listClusters(c *cli.Context) error {
	printClusters(c.Bool("all"))
	return nil
}

func getKubeConfig(c *cli.Context) error {
	sourcePath := fmt.Sprintf("%s:/output/kubeconfig.yaml", c.String("name"))
	destPath, _ := getClusterDir(c.String("name"))
	cmd := "docker"
	args := []string{"cp", sourcePath, destPath}
	if err := run(false, cmd, args...); err != nil {
		log.Fatalf("FAILURE: couldn't get kubeconfig for cluster [%s] -> %+v", c.String("name"), err)
		return err
	}
	fmt.Printf("%s\n", path.Join(destPath, "kubeconfig.yaml"))
	return nil
}

func main() {

	app := cli.NewApp()
	app.Name = "k3d"
	app.Usage = "Run k3s in Docker!"
	app.Version = "v0.1.1"
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
			Action: func(c *cli.Context) error {
				log.Print("Checking docker...")
				cmd := "docker"
				args := []string{"version"}
				if err := run(true, cmd, args...); err != nil {
					log.Fatalf("Checking docker: FAILED")
					return err
				}
				log.Println("Checking docker: SUCCESS")
				return nil
			},
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
					Usage: "Mount a volume into the cluster node (Docker notation: `source:destination`",
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
			},
			Action: createCluster,
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
			Action: deleteCluster,
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
			Action: stopCluster,
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
			Action: startCluster,
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
			Action: listClusters,
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
			Action: getKubeConfig,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(verbose bool, name string, args ...string) error {
	if verbose {
		log.Printf("Running command: %+v", append([]string{name}, args...))
	}
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
