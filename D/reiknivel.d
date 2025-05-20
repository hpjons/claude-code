import std.stdio;

void main(){

	int svar, tala1, tala2;
	
	writeln("1. Samlagning\n");
	writeln("2. Frádráttur\n");
	writeln("3. Margföldun\n");
	writeln("4. Deiling\n");

	writeln("Veldu : ");
	int svar = stdin.readln(svar);

	writeln("Sláðu inn tölu 1: ");
	int tala1 = stdin.readln(tala1);
	writeln("Sláðu inn tölu 2: ");
	int tala2 = stdin.readln(tala2);

	switch(svar){
		default:
			throw new Exception("Eitthvað fór úrskeiðis!");
		case 1:
			writeln("Summan er", tala1 + tala2);
			break;
		case 2: writeln("Mismunurinn er", tala1 - tala2);
			break;
		case 3: writeln("Margfeldið er", tala1 * tala2);
			break;
		case 4: writeln("Kvótinn er", tala1 / tala2);
			break;
	}
}

