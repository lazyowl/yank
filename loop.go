package main

import (
	"fmt"
	"lanfile/network/client"
	"lanfile/network/server"
	"lanfile/network/message"
	"lanfile/managers/config"
	"lanfile/managers/file_control"
	"lanfile/ui"
	"encoding/json"
	"log"
	"time"
)

const (
	LIST = 0
	LIST_REPLY = 1
)

type HighMessage struct {
	Cmd int
	Files []*file_control.MyFile
	Source string
}

func (m HighMessage) Serialize() string {
	b, _ := json.Marshal(m)
	return string(b)
}

func Deserialize(b string) HighMessage {
	var msg HighMessage
	json.Unmarshal([]byte(b), &msg)
	return msg
}

func main() {
	fmt.Println("Hello!")

	// read configuration
	err_read := config.Read_config()
	if err_read != nil {
		log.Fatal(err_read)
	}

	// client
	ch_c := make(chan message.Response)
	c, err_c := client.NewClient(ch_c)
	if err_c != nil {
		log.Fatal(err_c)
	}
	go c.ListenUnicast()
	go c.ListenMulticast()

	// server
	ch_s := make(chan message.Response)
	s, err_s := server.NewServer(ch_s)
	if err_s != nil {
		log.Fatal(err_s)
	}
	go s.Listen()

	// stdin IO
	io_struct := ui.NewIO()
	go io_struct.StdinListen()

	fc := file_control.File_controller{}
	fc.Init()

	kill_server := make(chan bool)
	go func() {
		for {
			select {
				case server_msg := <-s.Recv_ch: {
					high_msg := Deserialize(server_msg.Msg.Value)
					if high_msg.Cmd == LIST {
						response := HighMessage{LIST_REPLY, fc.List_local_files(), config.Config.Name}
						msg :=message.CreateMessage(0, response.Serialize())
						s.SendUnicast(server_msg.From, msg)
					}
				}
				case <-kill_server: {
					break
				}
			}
		}
	}()


	/* client loop */
	for {
		inp := <-io_struct.IO_chan
		switch inp.Op {
			case ui.LIST_CMD: {
				m := HighMessage{LIST, nil, config.Config.Name}
				c.SendMulticast(message.CreateMessage(0, m.Serialize()))
				timer := time.NewTimer(time.Second * 5)
				replies := []HighMessage{}
				reply_wait := true
				for {
					select {
						case client_msg := <-c.Recv_ch: {
							high_msg := Deserialize(client_msg.Msg.Value)
							if high_msg.Cmd == LIST_REPLY {
								replies = append(replies, high_msg)
							}
						}
						case <-timer.C: {
							reply_wait = false
							break
						}
					}
					if !reply_wait {
						break
					}
				}

				table := Merge(fc, replies)
				table.Display()
			}
			case ui.GET_CMD: {
			}
			case ui.LIST_LOCAL_CMD: {
				l := fc.List_local_files()
				for _, f := range l {
					fmt.Println(f)
				}
			}
		}

		io_struct.IO_cmd_complete <- true
	}
}

func Merge(fc file_control.File_controller, replies []HighMessage) ui.Stdout_table {
	// merge
	m := make(map[string][]*file_control.MyFile)
	users := make(map[string][]string)
	for _, r := range replies {
		for _, f := range r.Files {
			m[f.Full_hash] = append(m[f.Full_hash], f)
			users[f.Full_hash] = append(users[f.Full_hash], r.Source)
		}
	}

	table := ui.Stdout_table{}

	for k, v := range m {
		r := ui.Stdout_record{}
		r.Full_hash = k
		merged_hash_bit_vector := &file_control.Bit_vector{}
		var representative_file *file_control.MyFile
		for _, f := range v {
			r.Names = append(r.Names, f.Name)
			merged_hash_bit_vector.Bit_vector_or(f.Hash_bit_vector)
			representative_file = f
		}
		r.Percent_complete = merged_hash_bit_vector.Percent_set(representative_file.Get_num_blocks())
		local_copy := fc.Get_file_from_hash(k)
		if local_copy == nil {
			r.Percent_local = 0
			r.Max_complete = r.Percent_complete
		} else {
			r.Percent_local = local_copy.Percent_complete()
			merged_hash_bit_vector.Bit_vector_or(local_copy.Hash_bit_vector)
			r.Max_complete = merged_hash_bit_vector.Percent_set(representative_file.Get_num_blocks())
		}
		r.Users = users[k]
		table.Records = append(table.Records, r)
	}
	return table
}
