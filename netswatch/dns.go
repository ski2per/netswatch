// Copyright 2015 flannel authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package netswatch

import (
	"fmt"

	consul "github.com/hashicorp/consul/api"
)

type NWService struct {
	ID      string
	Address string
	Tags    []string
	Name    string
}

type DNSRegistry struct {
	Endpoint string
	Token    string
}

func (dnsr *DNSRegistry) ListService() {
	cli, err := consul.NewClient(&consul.Config{
		Address: dnsr.Endpoint,
	})
	if err != nil {
		panic(err)
	}

	agent := cli.Agent()
	svcs, err := agent.Services()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", svcs)
}

func (dnsr *DNSRegistry) RegisterSvc() {
	cli, err := consul.NewClient(&consul.Config{
		Address: dnsr.Endpoint,
		Token:   dnsr.Token,
	})
	if err != nil {
		panic(err)
	}
	agent := cli.Agent()

	var svc consul.AgentServiceRegistration

	svc.ID = "Demo2"
	svc.Name = "heheda"

	regErr := agent.ServiceRegister(&svc)
	if regErr != nil {
		panic(regErr)
	}

}
