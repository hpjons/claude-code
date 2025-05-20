import std.stdio;

void main(){
    for (int i = 0; i < 1000_000; i++){
        write(i, "...");
    }
    writeln("DONE");
}
