package adapter

import (
	"golang.org/x/net/context"
	"io"

	app "github.com/BrobridgeOrg/gravity-adapter-native/pkg/app"
	pb "github.com/BrobridgeOrg/gravity-api/service/adapter"
)

var SendEventSuccess = pb.SendEventReply{
	Success: true,
}

type Service struct {
	app      app.App
	Incoming chan []byte
}

func NewService(a app.App) *Service {

	service := &Service{
		app: a,
	}

	return service
}

func (service *Service) pushAsync(eventName string, payload []byte) {
	//TODO parsing payload struct
	service.Incoming <- payload
}

func (service *Service) SendEvent(ctx context.Context, in *pb.SendEventRequest) (*pb.SendEventReply, error) {

	// TODO
	/*
		controller := service.app.GetController()

		return &pb.GetClientCountReply{
			Count: controller.GetClientCount(),
		}, nil
	*/
	return &SendEventSuccess, nil

}

func (service *Service) SendEventStream(stream pb.Adapter_SendEventStreamServer) error {

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}
		/*
		   id := atomic.AddUint64((*uint64)(&counter), 1)

		   if id%1000 == 0 {
		           log.Info(id)
		   }
		*/

		//
		service.pushAsync(in.EventName, in.Payload)
	}
}
