// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
import * as vscode from 'vscode';
import * as child_process from 'child_process';

// this method is called when your extension is activated
// your extension is activated the very first time the command is executed
export function activate(context: vscode.ExtensionContext) {

	// Use the console to output diagnostic information (console.log) and errors (console.error)
	// This line of code will only be executed once when your extension is activated
	console.log('Congratulations, your extension "ytt-lint" is now active!');
	let timeout: NodeJS.Timer | undefined = undefined;

	// The command has been defined in the package.json file
	// Now provide the implementation of the command with registerCommand
	// The commandId parameter must match the command field in package.json
	let disposable = vscode.commands.registerCommand('extension.helloWorld', () => {
		// The code you place here will be executed every time your command is executed

		// Display a message box to the user
		vscode.window.showInformationMessage('Hello World!');
		//vscode.window.showInformationMessage(`activeEditor.document.languageId: ${activeEditor.document.languageId}`);

	});

	let diagnosticCollection = vscode.languages.createDiagnosticCollection('ytt-lint');
	
	context.subscriptions.push(disposable);
	context.subscriptions.push(diagnosticCollection);

	function lint() {
		if (activeEditor == null) {
			return;
		}
		if (activeEditor.document.languageId != "yaml") {
			return;
		}
		let doc = activeEditor.document;

		// TODO: don't use '-f -' if file is saved
		let yaml = doc.getText();

		vscode.window.showInformationMessage('Running lint now!');

		diagnosticCollection.clear();
		let diagnosticMap: Map<string, vscode.Diagnostic[]> = new Map();

		// TODO: use spwan and then stream
		let linter = child_process.execFile('/home/d060677/go/src/github.com/k14s/ytt/ytt-lint', ['-f', '-', '-o', 'json'], (error, stdout, stderr) => {
			vscode.window.showInformationMessage(stdout);
			let errors = JSON.parse(stdout);
			
			errors.forEach((error: { pos: string; msg: string; }) => {
				let [file, l] = error.pos.split(":");
				let lineNum = parseInt(l) - 1;
				//let canonicalFile = vscode.Uri.file(file).toString();
				let canonicalFile = doc.uri.toString();

				let line = doc.lineAt(lineNum);
				let start = line.firstNonWhitespaceCharacterIndex;
				let end = line.range.end.character;
				//let range = line.range;
				//range.start = line.firstNonWhitespaceCharacterIndex;

				let range = new vscode.Range(lineNum, start, lineNum, end);
				let diagnostics = diagnosticMap.get(canonicalFile);
				if (!diagnostics) { diagnostics = []; }
				diagnostics.push(new vscode.Diagnostic(range, error.msg /*TODO: , error.severity*/));
				diagnosticMap.set(canonicalFile, diagnostics);
			});
			diagnosticMap.forEach((diags, file) => {
				diagnosticCollection.set(vscode.Uri.parse(file), diags);
			});

		});
		linter.stdin?.write(yaml);
		linter.stdin?.end();
	}

	function triggerUpdateDecorations() {
		if (timeout) {
			clearTimeout(timeout);
			timeout = undefined;
		}
		timeout = setTimeout(lint, 500);
	}

	let activeEditor = vscode.window.activeTextEditor;

	if (activeEditor) {
		triggerUpdateDecorations();
	}

	vscode.window.onDidChangeActiveTextEditor(editor => {
		activeEditor = editor;
		if (editor) {
			triggerUpdateDecorations();
		}
	}, null, context.subscriptions);

	vscode.workspace.onDidChangeTextDocument(event => {
		if (activeEditor && event.document === activeEditor.document) {
			triggerUpdateDecorations();
		}
	}, null, context.subscriptions);
}

// this method is called when your extension is deactivated
export function deactivate() {}
