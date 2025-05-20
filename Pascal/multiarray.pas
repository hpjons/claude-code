program ex2dimarray;
const
	maxnum = 1000;
var		
	a: array [0..maxnum, 0..maxnum] of integer;
	i,j : integer;  
begin  
 
   maxnum = 1000

   for i:=0 to maxnum do
   begin
      for j:=0 to maxnum do
         a[i,j]:= i * j;  
   end;
   for i:=0 to maxnum do
   begin  
      for j:=0 to maxnum do  
         write(a[i,j]:2,' ');  
      writeln;  
   end;  
end.
