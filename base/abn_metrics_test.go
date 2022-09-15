package base

import (
	"context"
	"io/ioutil"
	"net"
	"testing"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	pb "github.com/iter8-tools/iter8/abn/grpc"
	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestABNMetricsTask(t *testing.T) {

	k8sclient.Client = *k8sclient.NewFakeKubeClient(cli.New())
	byteArray, _ := ioutil.ReadFile(CompletePath("../../testdata", "abninputs/readtest.yaml"))
	s, _ := k8sclient.Client.Typed().CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app",
			Namespace: "default",
		},
		StringData: map[string]string{"versionData.yaml": string(byteArray)},
	}, metav1.CreateOptions{})
	s.ObjectMeta.Labels = map[string]string{"foo": "bar"}
	k8sclient.Client.Typed().CoreV1().Secrets("default").Update(context.TODO(), s, metav1.UpdateOptions{})

	task := &collectABNMetricsTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectABNMetrics),
		},
		With: ABNMetricsInputs{
			Application: "app",
		},
	}

	exp := &Experiment{
		Spec:   []Task{task},
		Result: &ExperimentResult{},
	}

	exp.initResults(1)

	err := task.run(exp)
	assert.NoError(t, err)

	// any other assertions
}

func TestFoo(t *testing.T) {
	_, teardown := setup(t)
	defer teardown()

	task := &collectABNMetricsTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectABNMetrics),
		},
		With: ABNMetricsInputs{
			Application: "app",
		},
	}
	exp := &Experiment{
		Spec:   []Task{task},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)

	err := task.run(exp)
	assert.NoError(t, err)
}

func setup(t *testing.T) (*pb.ABNClient, func()) {
	k8sclient.Client = *k8sclient.NewFakeKubeClient(cli.New())
	// populate watcher.Applications with test applications
	abnapp.Applications.Clear()
	a, err := yamlToApplication("default/application", "../../testdata", "abninputs/readtest.yaml")
	assert.NoError(t, err)
	abnapp.Applications.Put(a)

	// start server
	lis, err := net.Listen("tcp", ":0")
	assert.NoError(t, err)

	serverOptions := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(serverOptions...)
	pb.RegisterABNServer(grpcServer, newServer())
	go grpcServer.Serve(lis)

	// setup client
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(lis.Addr().String(), opts...)
	assert.NoError(t, err)

	c := pb.NewABNClient(conn)

	// return client and teardown function to clean up
	return &c, func() {
		grpcServer.Stop()
		lis.Close()
		conn.Close()
	}
}

type abnServer struct {
	pb.UnimplementedABNServer
}

// newServer returns a new gRPC server
func newServer() *abnServer {
	return &abnServer{}
}
