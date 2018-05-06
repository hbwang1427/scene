package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"syscall"

	pb "github.com/aitour/scene/serverpb"
	"golang.org/x/net/trace"

	"github.com/BurntSushi/toml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	_ "image/jpeg"
	_ "image/png"
)

var (
	conf                = flag.String("conf", "service.toml", "Specify the config file")
	config              *Config
	MissingWebHostError = errors.New("webhost config missing")
)

type Config struct {
	Grpc struct {
		Bind        string
		Cert        string
		Key         string
		TraceEnable bool
		TraceBind   string
	}

	Web struct {
		Host string
	}
}

func parseConfig(conf string) (*Config, error) {
	var c Config
	content, err := ioutil.ReadFile(conf)
	if err != nil {
		return nil, err
	}
	if _, err = toml.Decode(string(content), &c); err != nil {
		return nil, err
	}
	if len(c.Web.Host) == 0 {
		return nil, MissingWebHostError
	}
	c.Web.Host = strings.TrimRight(c.Web.Host, "/")

	return &c, nil
}

// server is used to implement serverpb.AuthServer
type authserver struct{}

// Authenticate implements serverpb.AuthServer.Authenticate
func (s *authserver) Authenticate(ctx context.Context, in *pb.AuthRequest) (*pb.AuthResponse, error) {
	var response = &pb.AuthResponse{}
	if (in.Name != "test" || in.Password != "123") && in.Token != "this is an valid token" {
		response.Msg = "invalid name or password or invalid token"
	} else {
		response.Token = "this is another valid token"
	}
	return response, nil
}

//predictserver is used to implement serverpb.PredictServer
type predictserver struct{}

//PredictPhoto implements serverpb.PredictServer
func (s *predictserver) PredictPhoto(ctx context.Context, in *pb.PhotoPredictRequest) (*pb.PhotoPredictResponse, error) {
	var response = &pb.PhotoPredictResponse{}

	//ioutil.WriteFile("upload.jpg", in.Data, 0666)

	// m, _, err := image.Decode(bytes.NewReader(in.Data))
	// if err != nil {
	// 	return nil, err
	// }
	// bounds := m.Bounds()
	// var histogram [16][4]int
	// for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
	// 	for x := bounds.Min.X; x < bounds.Max.X; x++ {
	// 		r, g, b, a := m.At(x, y).RGBA()
	// 		// A color's RGBA method returns values in the range [0, 65535].
	// 		// Shifting by 12 reduces this to the range [0, 15].
	// 		histogram[r>>12][0]++
	// 		histogram[g>>12][1]++
	// 		histogram[b>>12][2]++
	// 		histogram[a>>12][3]++
	// 	}
	// }

	// var b bytes.Buffer
	// fmt.Fprintf(&b, "%-14s %6s %6s %6s %6s\n", "bin", "red", "green", "blue", "alpha")
	// for i, x := range histogram {
	// 	fmt.Fprintf(&b, "0x%04x-0x%04x: %6d %6d %6d %6d\n", i<<12, (i+1)<<12-1, x[0], x[1], x[2], x[3])
	// }
	// response.Text = b.String()

	audiourl := fmt.Sprintf("%s/assets/audio/sample_0.4mb.mp3", config.Web.Host)
	language := in.Language
	if language == "zh" {
		response.Results = []*pb.PhotoPredictResponse_Result{&pb.PhotoPredictResponse_Result{
			Text:      "故人西辞黄鹤楼，烟花三月下扬州，孤帆远影碧空尽，唯见长江天际流",
			ImageUrl:  fmt.Sprintf("%s/assets/imgs/c1.jpg", config.Web.Host),
			AudioUrl:  audiourl,
			AudioSize: 443926,
			AudioLen:  27,
		}, &pb.PhotoPredictResponse_Result{
			Text: `君不见，黄河之水天上来⑵，奔流到海不复回。
君不见，高堂明镜悲白发，朝如青丝暮成雪⑶。
人生得意须尽欢⑷，莫使金樽空对月。
天生我材必有用，千金散尽还复来。
烹羊宰牛且为乐，会须一饮三百杯⑸。
岑夫子，丹丘生⑹，将进酒，杯莫停⑺。
与君歌一曲⑻，请君为我倾耳听⑼。
钟鼓馔玉不足贵⑽，但愿长醉不复醒⑾。
古来圣贤皆寂寞，惟有饮者留其名。
陈王昔时宴平乐，斗酒十千恣欢谑⑿。
主人何为言少钱⒀，径须沽取对君酌⒁。
五花马⒂，千金裘，呼儿将出换美酒，与尔同销万古愁⒃`,
			ImageUrl: fmt.Sprintf("%s/assets/imgs/c2.jpg", config.Web.Host),
		}, &pb.PhotoPredictResponse_Result{
			Text: `明月出天山⑵，苍茫云海间。
长风几万里，吹度玉门关⑶。
汉下白登道⑷，胡窥青海湾⑸。
由来征战地⑹，不见有人还。
戍客望边邑⑺，思归多苦颜。 [1] 
高楼当此夜，叹息未应闲⑻。 [2] `,
			ImageUrl:  fmt.Sprintf("%s/assets/imgs/c3.jpg", config.Web.Host),
			AudioUrl:  audiourl,
			AudioSize: 443926,
			AudioLen:  27,
		}, &pb.PhotoPredictResponse_Result{
			Text: `红豆生南国⑵，春来发几枝⑶？
愿君多采撷⑷，此物最相思⑸。 [1] `,
			ImageUrl: fmt.Sprintf("%s/assets/imgs/220px-Buckman_Tavern_Lexington_Massachusetts.jpg", config.Web.Host),
		}, &pb.PhotoPredictResponse_Result{
			Text: `云想衣裳花想容， 春风拂槛露华浓。
若非群玉山头见， 会向瑶台月下逢。`,
			ImageUrl: fmt.Sprintf("%s/assets/imgs/250px-Minute_Man_Statue_Lexington_Massachusetts.jpg", config.Web.Host),
		}, &pb.PhotoPredictResponse_Result{
			Text: `一枝红艳露凝香，云雨巫山枉断肠。
借问汉宫谁得似？ 可怜飞燕倚新妆。`,
			ImageUrl: fmt.Sprintf("%s/assets/imgs/3178927_orig.jpg", config.Web.Host),
		}, &pb.PhotoPredictResponse_Result{
			Text: `名花倾国两相欢，长得君王带笑看。
解释春风无限恨，沉香亭北倚阑干。`,
			ImageUrl: fmt.Sprintf("%s/assets/imgs/vt.jpg", config.Web.Host),
		}, &pb.PhotoPredictResponse_Result{
			Text: `金樽清酒斗十千⑴，玉盘珍羞直万钱⑵。
　　停杯投箸不能食⑶，拔剑四顾心茫然。
　　欲渡黄河冰塞川，将登太行雪满山。
　　闲来垂钓碧溪上，忽复乘舟梦日边⑷。
　　行路难！行路难！多岐路，今安在⑸？
　　长风破浪会有时⑹，直挂云帆济沧海⑺`,
			ImageUrl: fmt.Sprintf("%s/assets/imgs/images3.jpeg", config.Web.Host),
		}}
	} else if language == "de" {
		//
	} else {
		//default en
		response.Results = []*pb.PhotoPredictResponse_Result{&pb.PhotoPredictResponse_Result{
			Text:      "Appel was a founding member of CoBrA, a short-lived post-War association of painters, writers, and poets, whose name is an acronym for Copenhagen, Brussels, and Amsterdam, the capital cities of the founders’ countries. Emphasizing spontaneity and directness, some members, such as Appel, based their work on children’s drawings and folk art, and the art of Paul Klee, often in bold colors.",
			ImageUrl:  fmt.Sprintf("%s/assets/imgs/c1.jpg", config.Web.Host),
			AudioUrl:  audiourl,
			AudioSize: 443926,
			AudioLen:  27,
		}, &pb.PhotoPredictResponse_Result{
			Text:     "dish",
			ImageUrl: fmt.Sprintf("%s/assets/imgs/c2.jpg", config.Web.Host),
		}, &pb.PhotoPredictResponse_Result{
			Text:      "building",
			ImageUrl:  fmt.Sprintf("%s/assets/imgs/c3.jpg", config.Web.Host),
			AudioUrl:  audiourl,
			AudioSize: 443926,
			AudioLen:  27,
		}, &pb.PhotoPredictResponse_Result{
			Text:     "220px-Buckman_Tavern_Lexington_Massachusetts",
			ImageUrl: fmt.Sprintf("%s/assets/imgs/220px-Buckman_Tavern_Lexington_Massachusetts.jpg", config.Web.Host),
		}, &pb.PhotoPredictResponse_Result{
			Text:     "250px-Minute_Man_Statue_Lexington_Massachusetts",
			ImageUrl: fmt.Sprintf("%s/assets/imgs/250px-Minute_Man_Statue_Lexington_Massachusetts.jpg", config.Web.Host),
		}, &pb.PhotoPredictResponse_Result{
			Text:     "3178927_orig",
			ImageUrl: fmt.Sprintf("%s/assets/imgs/3178927_orig.jpg", config.Web.Host),
		}, &pb.PhotoPredictResponse_Result{
			Text:     "vt",
			ImageUrl: fmt.Sprintf("%s/assets/imgs/vt.jpg", config.Web.Host),
		}, &pb.PhotoPredictResponse_Result{
			Text:     "images3.jpeg",
			ImageUrl: fmt.Sprintf("%s/assets/imgs/images3.jpeg", config.Web.Host),
		}}
	}

	if in.MaxLimits > 0 && len(response.Results) > int(in.MaxLimits) {
		response.Results = response.Results[:in.MaxLimits]
	}
	//time.Sleep(20 * time.Second)
	return response, nil
}

func createGrpcServer() (*grpc.Server, error) {
	//make credentials for grpc
	cert, err := tls.LoadX509KeyPair(config.Grpc.Cert, config.Grpc.Key)
	if err != nil {
		return nil, err
	}
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   0,
	})

	// creds, err := credentials.NewServerTLSFromFile(config.Grpc.Cert, config.Grpc.Key)
	// if err != nil {
	// 	return nil, fmt.Errorf("Failed to generate credentials %v", err)
	// }
	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.MaxSendMsgSize(20 * 1024 * 1024), //max send message size set to 20MB
		grpc.MaxRecvMsgSize(20 * 1024 * 1024), //max recv message size set to 20MB
	}

	//create grpc server
	grpc.EnableTracing = config.Grpc.TraceEnable
	s := grpc.NewServer(opts...)

	//register serverpb.AuthServer
	pb.RegisterAuthServer(s, &authserver{})
	pb.RegisterPredictServer(s, &predictserver{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	return s, nil
}

func main() {
	flag.Parse()
	var err error
	if config, err = parseConfig(*conf); err != nil {
		log.Fatal(err)
	}

	//create grpc server
	s, err := createGrpcServer()
	if err != nil {
		log.Fatal(err)
	}

	//startup grpc server
	log.Printf("starting grpc server on %s", config.Grpc.Bind)
	lis, err := net.Listen("tcp", config.Grpc.Bind)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	//startup trace server if wanted
	//visit /debug/requests in browser to trace requests
	if config.Grpc.TraceEnable {
		trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
			return true, true
		}
		log.Printf("visit tracing at: %s", config.Grpc.TraceBind)
		go http.ListenAndServe(config.Grpc.TraceBind, nil)
	}

	//setup signal handlers to handle signals sent to stop this process
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	sig := <-quit
	log.Printf("signal %s received. shutdown server ...", sig.String())

	s.GracefulStop()

	log.Println("service stopped")
}
