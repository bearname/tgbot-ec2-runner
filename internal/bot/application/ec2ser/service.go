package ec2ser

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jackc/pgx"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

const privPEM = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAloKHfIJbtNWYxDl44snEyQs96/q0RNlcXk56Sl0gVPX7aIM+
oMYnj7Roz6mcftBlMGABkHsWiwk6ZPvqPdq2eamTFSqYUXf5Yizc3OwCOuV8aVhL
KeTCAJqJHkhvivPonpZ96oE60Z6SEO5AVBDQEkaU5X8Q6cGK0WLSfr6JB/++Le9a
meZ4ah8p6+exO5PZgM+l/uQY3b56t+g0Ux/FjPyyh8fq0QNzYMPv2UIpL87W4pmy
05axEwZgw/LdmhnV6A8uNOwb9H/0LgYhXU7L1EzkuaUrs3P1z/8U5Xr3AMb2xMGk
7ufSjIxTtUBA7VLb7+gmqBYbLJ5mH++dIFFv8QIDAQABAoIBAD53bkP+1pD3FbBb
KhD2LPZD9C88nhT1IaECcT7r579bWDzIO/X+R+0cs2N8wbbgRx8MuZl9fJ239sRy
yYVigNl9x83BH/awFJDqjcAjM8m99STDwG5iwyH9AWvQJHtHQASR8TCi8gaOCZF9
ULWTRMhRAvs9AYgDC9pbaYGxEq5+khLBDh1ntmZD0bIu4bmHTUw21/r8+Ju6KKP6
2Ae4KogWGKEe1HksVvRVaLk83BPvjkhNY+1BxtJip/HxNLzBDQHmmZP0M709b9id
3ilMM/nL8oOLjEiBHJS++aQhY8SgRQz6dsYfWKKzFjAxfjzwrA8ctkMUcncZ2f+y
7gnpkQECgYEA3VdZJ2Lx9kyBs3hUx35Z5ZEurmwZVZchEaQoYdXOPlmPP323CaIp
wgZrwY4Cx82X9IvTorx7U2/ttyH6yT1KU//ikMJwfJiaSCjeoed828q/lc7VoAmR
k3XOh5N3Ak525GjeTAdhXUe8Odm0A+JhQEOixfMgMZB7DuzqIGjJ3/cCgYEArhPY
wHsd4lxQb+tvyglWzTOn2ss3TZWlzWLtCUeOfsP8gF+uMRHuBV/zUH190xH98N8/
UmZDO7cvmvt7oXugbpqWXtqPbYAth6vggc1daHORG5JIeDMYPzXKBEF/XSXtMduX
al37BmqmmwgMtHlV0K6ZDeKjfRPj7sBz47+9hVcCgYBx+2pQ6xY5RNrB5iSaewmm
O6Zzcf114xbHc+bLwwOrfglTo9SfZF/mp9HT4eKyq8Al0d/RfQhxRkF/PkNcYHCn
Yy10aHzu3NMOd+V2MLROp1ETv2ipOmQ6ML+Dd8AgcvGs3Agl1OMh2zAmBmM6YNi9
9FadY39fpsyIOh6zQ+M5rwKBgFv7BqvmIgtKpgINUFtcBe6VndsBR+6J5TsaS498
rPGUk2YtqHgrNi7G3WUpegO+XQUaoXXjrSDvkYr92Pyhu0rWCiCCsgi1Etm+Wvmb
TwDzF7iO2hYRQX2c6WrIRQkuEiAnHOHKWOqyDeibH0N5XXvP1fW9TI+5o9WzAUlV
NkovAoGBAK+s164m2yVAnuYZCkY8kLDOciMv6fdf5OHZ/tZEzxTkQKTWsjQjYRNd
f0hsy2V/kF2myobFu4d9CyXHvVlpsAOJUuEobqgQiouxW4/5xTMgnUkJDy5Usn11
FU04+SwWoi7kzfLg3wwINpNsk6p1PLjM/WtkdLrAva4xaB96QtAl
-----END RSA PRIVATE KEY-----`

type Service struct {
	svc  *ec2.EC2
	repo map[string]int
	pool pgx.ConnPool
}

func NewService(pool *pgx.ConnPool, sess *session.Session) *Service {

	s := new(Service)
	s.svc = ec2.New(sess)
	s.repo = make(map[string]int, 0)
	s.pool = *pool
	return s
}

func (s *Service) GetAvailableInstances() []Ec2Instance {
	result, err := s.svc.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		log.Println("Got an error retrieving information about your Amazon EC2 instances:", err)
		return nil
	}

	rdpFiles := s.getEc2Instances(result)

	return rdpFiles
}

func (s *Service) getEc2Instances(result *ec2.DescribeInstancesOutput) []Ec2Instance {
	instances := make([]Ec2Instance, 0)
	for _, r := range result.Reservations {
		for _, i := range r.Instances {
			inst := *i
			instanceId := *inst.InstanceId
			s.repo[instanceId] = 1

			if inst.Platform != nil && *inst.Platform == "windows" {
				name := ""
				for _, tag := range i.Tags {
					if *tag.Key == "Name" {
						name = *tag.Value
					}
				}

				instance := Ec2Instance{
					State:         *inst.State,
					Id:            *inst.InstanceId,
					Name:          name,
					PublicDnsName: *inst.PublicDnsName,
				}
				instances = append(instances, instance)
			}
		}
	}
	return instances
}

func (s *Service) GetRdpFile(instanceId string) string {
	result, err := s.getInstances(instanceId)
	if err != nil {
		log.Println("Got an error retrieving information about your Amazon EC2 instances:")
		log.Println(err)
		return ""
	}

	for _, r := range result.Reservations {
		for _, i := range r.Instances {
			if instanceId != *i.InstanceId {
				continue
			}

			data, err := s.svc.GetPasswordData(&ec2.GetPasswordDataInput{InstanceId: &instanceId})
			if err != nil {
				log.Println("DescribeKeyPairsRequest readalls:", err)
				return ""
			}

			passwordData := *data.PasswordData
			passwordData = strings.ReplaceAll(passwordData, "\n", "")
			passwordData = strings.ReplaceAll(passwordData, "\r", "")
			ioutil.WriteFile("prim.pem", []byte(privPEM), 0644)
			ioutil.WriteFile("pass.txt", []byte(passwordData), 0644)

			out, err := exec.Command("bash", "-c", "base64 -d pass.txt | openssl rsautl -decrypt -inkey prim.pem").Output()
			if err != nil {
				os.Remove("prim.pem")
				os.Remove("pass.txt")
				log.Println(err)
				return ""
			}
			os.Remove("prim.pem")
			os.Remove("pass.txt")
			password := string(out)

			rdpFile := "auto connect:i:1\nfull address:s:" + *i.PublicDnsName + "\nusername:s:Administrator\npassword " + password

			ioutil.WriteFile(instanceId+".rdp", []byte(rdpFile), 0644)
			return rdpFile
		}
	}
	return ""
}

func (s *Service) getInstances(id string) (*ec2.DescribeInstancesOutput, error) {
	var ids []*string
	ids = append(ids, &id)
	input := ec2.DescribeInstancesInput{
		InstanceIds: ids,
	}

	result, err := s.svc.DescribeInstances(&input)
	if err != nil {
		return nil, err
	}

	return result, nil
}
