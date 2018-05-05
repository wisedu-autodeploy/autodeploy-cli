rm -rf bin/darwin bin/linux bin/windows &&
mkdir -p bin/darwin bin/linux bin/windows &&

gox -osarch="linux/amd64 windows/amd64 darwin/amd64" -output="autodeploy_{{.OS}}_{{.Arch}}" &&

mv autodeploy_darwin* bin/darwin/autodeploy &&
mv autodeploy_linux* bin/linux/autodeploy &&
mv autodeploy_windows* bin/windows/autodeploy.exe &&

chmod +x bin/darwin/autodeploy