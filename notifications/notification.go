package notifications

import "github.com/pusher/pusher-http-go"

func SendToNotificationChannel(message string, channel string, event string) {
	pusherClient := pusher.Client{
		AppID:   "1278539",
		Key:     "c371bbb2fc2670c038f2",
		Secret:  "02798a8fe29f1772c771",
		Cluster: "ap1",
		Secure:  true,
	}

	data := map[string]string{"message": message}
	pusherClient.Trigger(channel, event, data)
}
