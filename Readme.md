GO tcp2com

- Socket with Golang : https://gist.github.com/luscas/5a08a9f9d42764749e9a0c0053c908c9
- Serial : https://github.com/tarm/serial

tcp <-> com

go를 활용한 tcp와 com 통신 간의 변환입니다.
산업현장에서 아직도 열심히 일하고 있는 시리얼기기를 활용하여 간단하게 작성해봅니다.

<img src="tcp.png" width="30%" height="30%">
tcp/ip 로 데이터를 보내면.
<img src="com.png" width="30%" height="30%">
해당 데이터가 터미널에서 수신됩니다.