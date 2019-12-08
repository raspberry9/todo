#url="https://jskoo.iptime.org:3334"
url="https://127.0.0.1:3334"

echo "가입을 처리 합니다."
curl -X POST -k -d '{"id":"koo@jsproj.com","password":"1111","name":"myname"}' $url/signup
echo .

echo "이메일 인증을 처리 합니다."
actkey=`mysql -uroot auth -e "select activationkey from users where id='koo@jsproj.com';" | tail -n 1`
curl -X GET -k https://127.0.0.1:3334/activation/$actkey
read

echo "로그인을 하여 토큰을 얻습니다."
curl -X POST -k -d '{"id":"koo@jsproj.com", "password":"1111"}' $url/login > /tmp/jwttoken.tmp 2> /dev/null
token=`cat /tmp/jwttoken.tmp | cut -d ":" -f 4 | cut -d "}" -f 1 | cut -d '"' -f 2`
echo .

echo "로그인한 유저를 운영자로 승격 시킵니다."
mysql -uroot auth -e "update users set type=-1 where id='koo@jsproj.com';"
mysql -uroot auth -e "update users set status=1 where id='koo@jsproj.com';"
echo .

echo "서버 설정을 다시 읽습니다."
curl -X GET -k -H "Authorization: Bearer $token" $url/reloadconfig
echo .

#echo "일정을 몇가지 추가 합니다."
#insert into todolist (owneruid, category, todo, detail, place, limittime, completetime) values (1, 'Personal', '엄마 심부름', '당근\n라면', '마트', 0, 0);

echo "탈퇴 합니다."
curl -X POST -k -H "Authorization: Bearer $token" $url/withdraw
echo .
