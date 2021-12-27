package zapgo

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func zapInit() error {
	// create a new docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	defer cli.Close()

	// pull the zap image from DockerHub
	image, err := cli.ImagePull(ctx, "docker.io/owasp/zap2docker-weekly", types.ImagePullOptions{})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, image)

	// // create a container
	// container, err := client.NewContainer(
	// 	ctx,
	// 	"owasp-zap",
	// 	containerd.WithImage(image),
	// 	containerd.WithNewSnapshot("owasp-zap", image),
	// 	containerd.WithNewSpec(oci.WithImageConfig(image)),
	// )
	// if err != nil {
	// 	return err
	// }
	// defer container.Delete(ctx, containerd.WithSnapshotCleanup)

	// // create a task from the container
	// task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	// if err != nil {
	// 	return err
	// }
	// defer task.Delete(ctx)

	// // make sure we wait before calling start
	// exitStatusC, err := task.Wait(ctx)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// // call start on the task to execute the redis server
	// if err := task.Start(ctx); err != nil {
	// 	return err
	// }

	// // sleep for a lil bit to see the logs
	// time.Sleep(3 * time.Second)

	// // kill the process and get the exit status
	// if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
	// 	return err
	// }

	// // wait for the process to fully exit and print out the exit status

	// status := <-exitStatusC
	// code, _, err := status.Result()
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("zap exited with status: %d\n", code)

	return nil
}
