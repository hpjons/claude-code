import std.stdio;
import std.file;
import std.datetime;
import std.string;

void main(){
    writeln("Skrifaðu niður minnisatriði:");
    string note = readln().strip();

    string timestamp = Clock.currTime.toSimpleString();
    string line = format("[%s] %s\n", timestamp, note);

    append("notes.txt", line);

    writeln("Note saved.");
    writeln("Show all notes? (y/n)");
    string response = readln().strip();

    if (response == "y"){
        writeln("\n--- All Notes ---");
        writeln(cast(string) readText("notes.txt"));
    } else {
        writeln("Goodbye.");
    }
}