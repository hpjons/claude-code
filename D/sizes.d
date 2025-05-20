import std.stdio;

void main()
{
    writeln("Type  \tBytes\tBits");
    writeln("byte  \t", byte.sizeof, " ----> ", byte.sizeof * 8, " ----> ", 2^^8 );
    writeln("ubyte \t", ubyte.sizeof, " ----> ", ubyte.sizeof * 8, " ----> ", 2^^8);
    writeln("short \t", short.sizeof, " ----> ", short.sizeof * 8);
    writeln("ushort\t", ushort.sizeof, " ----> ", ushort.sizeof * 8);
    writeln("int   \t", int.sizeof, " ----> ", int.sizeof * 8);
    writeln("uint  \t", uint.sizeof, " ----> ", uint.sizeof * 8, " ----> ", 2^^32);
    writeln(long.sizeof);
    writeln(ulong.sizeof);
}