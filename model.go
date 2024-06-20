package main

type Config struct {
	Imap     string
	Email    string
	Password string
	LastUID  string
}

type LastMessageInfo struct {
	CountMessage int64  `json:"count_message"`
	LastUID      string `json:"last_uid"`
}
