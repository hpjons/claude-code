import std.process;

void main(){
    auto result = execute(["zenity", "--info", "--text=Hello from D!"]);
}