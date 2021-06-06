
# Send a random reward to baseline or candidate
send_reward () {
    R=$((1 + $RANDOM % 10))
    if [ $R -gt 6 ]
    then
	echo "Sending reward to baseline!"
	curl -d '{"reward": 1}'    -X POST http://172.18.255.1/seldon/ns-baseline/iris/api/v1.0/feedback    -H "Content-Type: application/json"
    else
	echo "Sending reward to candidate!"
	curl -d '{"reward": 1}'    -X POST http://172.18.255.1/seldon/ns-candidate/iris/api/v1.0/feedback    -H "Content-Type: application/json"
    fi    
}

# Send a request to A/B test
send_request () {
   curl -d '{"data": {"ndarray":[[6.8,2.8,4.8,1.4]]}}'    -X POST http://172.18.255.1/api/v1.0/predictions    -H "Content-Type: application/json" -HHost:iris.example.com    
}

i=0
while true; do
    let i=i+1
    if ! ((i % 4)); then
	send_reward
    fi
    send_request
    sleep 0.2
done
