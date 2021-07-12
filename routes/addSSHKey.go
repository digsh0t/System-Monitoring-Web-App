package routes

import (
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func AddSSHKey(w http.ResponseWriter, r *http.Request) {
	//Dummy value
	keyName := "id_rsa"
	privateKey := `-----BEGIN OPENSSH PRIVATE KEY-----
	b3BlbnNzaC1rZXktdjEAAAAACmFlczI1Ni1jdHIAAAAGYmNyeXB0AAAAGAAAABAZzsCYyY
	0BrKXHvoHk3UWvAAAAEAAAAAEAAAIXAAAAB3NzaC1yc2EAAAADAQABAAACAQC8vK3SUEc/
	hKFdrvhgLfSNkye3DVYZm01fgHNdFuaoKhGl8z4+8vKWN+RItzxe+SEnCuPXmY+5tZCHKY
	SY/0rRfJ+b+AxU7/ezqsXc7J6l7Hc0K7h7KBEsHU152kDW30XXvaSynU9Muoq2Vi9aIer4
	qFd5NKEMfPaXt+imVvMhU6t/8e0Xlr6d63BTmJdE9Rz0ty3IxHzoi0jNQlMPkvlpao/Syj
	98V7t4BHwNLdi9PX2u6J1NuDW64rw/1XRIOsRFoF2mqqSTAYKUA7v+0dYO9ZIZ2LgMlp/h
	ts6jYLio5DEUUZJ3LHMKF2WHQ3zCEQJ+GlOMptAfx1JKOP0O4XlI8AF8/d5RGjtPVy10GB
	1oS6Eaembhwd0ovtoqKPGi+7hek8d/U1IY4B4D3WFtd+UZdmr+WiL+fvi8pI/T0NeA22Us
	0uPXJBbovcKcPVX5U4iGw62XviiSUBzqMBrvMogrMPlX+oiUXHP1hiuAtKQk4sTg76iHpo
	IZMkbRKzclTQzST+lxxLAQTKsWoN1gwqL7EbXOj4R64xDHr3W9xiiqe9/kZg7R12rmWAKf
	SBF5x0MpyYtIigSl0VAtI6xsQIVOBs9nLNL2zyy6xo1TXoLHEBlAh7PctOYdUY72h+HHPw
	yegQoYxUSoQ4BCtEpq/BcJFAfBkSXXl6iD805WKQ/mZwAAB1ANC6TdISg1MA+/0Zjnd3JG
	/dqJRQZUZ/ni6Y53FSodgfz1DoX8nSBVlxO+sv+Ify9TblJSMr6jvoauWELoVPF1TS1fX3
	0GXIDPPaqg2OZ/zxrhDfbVQw2CE4uasLcJ0dyfXNJxTOqIwlwWNZMRLECjDtDyfa5Yat4l
	SRV7sObag6JnWe/ABbVqj8A325eBTlx8q8di+e4qff8p/7+DeFc3p9QOtIU5au3tfEgZVh
	ezx6zP9De1bFD3pPE2SgVUJy21hCCc2x1+k93F67WeGzhWlmZPV4uuyECxvt00wpPCdail
	pLsqQrpzPYAyqmyXICWw/BCU1d43LO4axnunNZaLPVcnT+KXOpoFRITxUi4inLRlf0XPZV
	RkeUC5/ppSgeOTRlhViQL4SgcFEtn3MQJMlek+gUzp8QAoMgsU6OFc3FngBVkjQF9vDm55
	3EXUV1s4VbneHTe7KTAMZdGLYl47uEb5I9JURssicH2f1K6z2aRWxy98FZ4yJcnG6dyom7
	r5X5JovpCo2WZ5DP6+quW6r0kQz0sAJG0Qp3+rXNS5AlUmRqR8er0B3TCBzwOwIgI7AcG1
	Eq6tXkWcqS/fA3qwK9mVMyvc0bkHiJ+XBhFClxoOnw6+VeIQvIZk6oPNHuFTMB8/HvE7y6
	MJXXtZKgOIKbNOLKyiBiBuAcczf7QbGLhrXoMT6FHNtJsSUvOrasWg5ityRkk914ZFYM3h
	+erU9ejk+h+N2Ag1s2YxIMr17U1h5la9ceZt6+55sv4rHqzqA01YHl0RPhlxKRCvQmESSU
	Y9iLNFiQnFLp5bqkhGbNpkcCM2c/+NNpwcc7KzjlO86DJKSueLPdB88wkyjuHuzaUB5yn1
	SJsKqCjs2TACuCd0VYju0y9AuJJa1Im+0eXsO8GeMQ59DaCNR6mgwgUnKQXjV9Ovxbqk+1
	7oQhMdYipVmg5hgoVr4YEpQyirORyPpooLTrH9pLzlOTXmvx1RPkiMNLqSs2h1+R0OEPSd
	MD1moJm7oXaPaqqMfuNkq93H0VLTapKn1eFeSd7aou3naZzsxj27wKG/wzHUu7ikEEW2V2
	+W3llDA8O6LFJvsuY719XoJRdySRMGywHOWTvfWyPeo6KMUdFGBmEGcDrGL0nLw+RW08g8
	m0bhtPPaIr9URpLlRlvgABq8An3vtW9vyp3lZ9GmGGD4v8gkvPgdJxulqF4eBXYW0AVg0/
	G8XOrCqeMWggcfPeGjNRlw2EoLkGF8ux/ZBdJcMbxyZi/AaU8j6oTRuQC7fjU17K+2orvy
	i3audh2NjrVqdeqVMUtVpDXiosYmTiBAjcsv4BKK4+skyxAWi+v/oieQd4b5WGbFNiVQCf
	bf2vlUmEKx5nN1rQnJsRp26UuM9hBzV7DMqqukPkIi9SefoYjRzQ1T+b7cmVslfbPoqV+l
	GQIqTouWTo6D4n9GLVqBeq0cGzIbwIZPSqvDk2sIj/YbeGWV416YHVbXMkVdl1xdhX5Kt0
	XEGiOkl15bJAPXkA3gRFQEGV4CjdZsSN1CFzvWvtknZJlw0/JqVxAmp+GoGxhukrbusNgX
	WKRXABH3GjVB+96D93FwJwHG1LT7gXLv8VUzzO3JtLYvOvHOwqVQC6A/dql/S29SnNOfNT
	B3RrL03SHbNe2UoJUJwmEUZeE+k65gQ5TOB6ZgCTM7As43pYt/956FSCBJ2e4fWoQG7vvf
	TF00AZ8+RWf2vQJwkkJ6qKp5f5ZznARuxlEgITRO2XZPEQPtEqF+Kgg4XNSTVV/OMCz1IK
	DcemQOe1qOoDdp5eQmhXR+6sitE+xD99x2cpuEZ/74EV4V4obR1lEanwhixIx5bUFqV/2E
	GxmrnRoIcJZwecUBeCLAsd/qhSLJkV6GxYlS3syMC6AeRK6phig+zrVyaxzq89jGZ7HT6z
	/WNATpiVWEoFhsoE6lqRw6xNCx9dwGz8g7T4n00KwWDzcW4OBL+DDdffboSYiOmThYdcAF
	A4S3m+4Am1KVoAy0pj0nBrTG8HVXg4XUk6n+7+CmqN6WpHmiqMzjX48x6j5XnCUycDKhar
	nN28W+RlNuP77GbqRk1Ptu8HzzScQBPZCEKBetKGg3MpXBX7ZAxenRG0oJCVBWzFtbi9ph
	nN+hWtwce8d/5zmgUhnqg0t7EARHOgCNIFWV8AtBfQTG1LDn5TLwRYqkDXAvmH8FbWpxn9
	Cq33dV5DTNgDOuoZjIXAPaof+hxRI1xhXkz4RlCjut672tur0bU0oFczt5J8ogO/E33mrd
	/HRvpi3LrF89rlqqy++wgPE8HJ1He/9O+s+d6V0wwh46byYkvPEtJAjSSR0Z2lzLuEFzTY
	SVxXZfikQz9cN5pvYqRTrE/znA35Q36XfqIX6OlEd/WZ/LmW8sq3EYyrGGYYKje8841bGO
	/aZZyUCGxcSH3v5R5Lm/13JCkIYz+170xU8clNhdp85YIPCo57pkMYm2VAbp2y9Bu+351B
	dO0gBAOVHOHIPPsQzbPdcsBQI=
	-----END OPENSSH PRIVATE KEY-----`
	creatorId := 1

	var sshKey models.SSHKey
	sshKey.KeyName = keyName
	sshKey.PrivateKey = models.AESEncryptKey(privateKey)
	sshKey.CreatorId = creatorId

	status, err := sshKey.InsertSSHKeyToDB()
	returnJson := simplejson.New()
	returnJson.Set("Status", status)
	statusCode := http.StatusBadRequest
	if err != nil {
		returnJson.Set("Error", err.Error())
	} else {
		returnJson.Set("Error", err)
		statusCode = http.StatusCreated
	}
	utils.JSON(w, statusCode, returnJson)
}
