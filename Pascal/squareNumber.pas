program test;
var
	a, x: integer;
	ans: char;
procedure square(varToProcedure: integer; var varFromProcedure: integer);
begin
	varFromProcedure := varToProcedure * varToProcedure;
end;

begin
	ans := 'j';
	repeat
		writeln('Enter number : ');
		readln(a);
		square(a, x);
		writeln('Squared = ', x);
		writeln('Again (j/n)');
		readln(ans)
	until ans = 'n';
end.
