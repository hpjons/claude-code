import std.stdio;
import std.file;
import std.string;
import std.conv;
import std.math;

void main()
{

    auto input(T)() => to!T(strip(readln()));


    //write("A? : ");
    //auto A = input!double;
    //write("B? : ");
    //auto B = input!double;
    //write("C? : ");
    //auto C = input!double;
    auto A = -2.34;
    auto B = 3.12;
    auto C = -1.43;

    auto D = pow(B,2) - 4 * A * C;

    if (D < 0) {
        writeln("Only complex solutions");
    }
    else if (D == 0) {
        writeln("One solution");
    }
    else {
        writeln("Two solutions");
    }

}