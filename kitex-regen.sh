cd ./rpc-server || exit
echo "Re-generating KiteX code..."
kitex -module "github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server" -service imservice ../idl_rpc.thrift
cp -r ./kitex_gen ../http-server # copy kitex_gen to http-server