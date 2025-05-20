program SimpleGUI;

uses
	Forms, StdCtrls, Interfaces;

var
	Form: TForm;
	Button: TButton;
begin
	Application.Initialize;
	Application.CreateForm(TForm, Form);

	Button := TButton.Create(Form);
	Button.Parent := Form;
	Button.Caption := 'Click me';
	Button.Left := 50;
	Button.Top := 50;
	Button.OnClick := @Application.Terminate;

	Form.Caption := 'Hello Lazarus';
	Form.Width := 300;
	Form.Height := 200;

	Application.Run;
end.
