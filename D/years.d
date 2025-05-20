import std.stdio;

void main()
{
	auto counter = 0;

	for (int i = 1977; i < 2025; i++){
		if (i == 2012) {
			writeln(counter, " ", i, " Rúrik");
			counter = counter + 1;
		} else {
			writeln(counter," ", i);
			counter = counter + 1;
		}
	}
}
