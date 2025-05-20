import std.stdio;
import std.string;
import std.process;

void main(){
    auto result = execute(["zenity", "--entry", "--title=Nafnið þitt", "--text=Hvað heitir þú?"]);

    if (result.status == 0){
        writeln("Hæ hó, ", result.output.strip(), "!");
    } else {
        writeln("Notandi hætti við eða villa kom upp!");
    }
    
}