import { execFile } from "child_process";
import * as path from "path";
import { promisify } from "util";
import * as vscode from "vscode";
import {
  Executable,
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
} from "vscode-languageclient/node";

const execFileAsync = promisify(execFile);

let client: LanguageClient | undefined;
// Track the last active text editor for returning focus after webview clicks
let lastActiveTextEditor: vscode.TextEditor | undefined;

export function activate(context: vscode.ExtensionContext) {
  // Start LSP client
  startLanguageClient(context);

  // Track the last active text editor (before webview takes focus)
  context.subscriptions.push(
    vscode.window.onDidChangeActiveTextEditor((editor) => {
      if (editor) {
        lastActiveTextEditor = editor;
      }
    })
  );

  // Register visualize file command
  const visualizeCommand = vscode.commands.registerCommand(
    "twf.visualize",
    async (uri?: vscode.Uri) => {
      // If called from explorer context menu, use the URI
      if (uri) {
        await WorkflowVisualizerPanel.createOrShowForFile(context.extensionUri, uri.fsPath);
        return;
      }

      // Otherwise use active editor
      const editor = vscode.window.activeTextEditor;
      if (!editor || editor.document.languageId !== "twf") {
        vscode.window.showWarningMessage("Please open a .twf file to visualize");
        return;
      }
      await WorkflowVisualizerPanel.createOrShowForFile(context.extensionUri, editor.document.uri.fsPath);
    }
  );

  // Register visualize folder command
  const visualizeFolderCommand = vscode.commands.registerCommand(
    "twf.visualizeFolder",
    async (uri?: vscode.Uri) => {
      let folderPath: string | undefined;

      if (uri) {
        folderPath = uri.fsPath;
      } else {
        // Prompt user to select a folder
        const folders = await vscode.window.showOpenDialog({
          canSelectFiles: false,
          canSelectFolders: true,
          canSelectMany: false,
          title: "Select folder containing .twf files",
        });
        if (folders && folders.length > 0) {
          folderPath = folders[0].fsPath;
        }
      }

      if (!folderPath) {
        return;
      }

      // Find all .twf files in the folder
      const pattern = new vscode.RelativePattern(folderPath, "**/*.twf");
      const uris = await vscode.workspace.findFiles(pattern);

      if (uris.length === 0) {
        vscode.window.showWarningMessage("No .twf files found in the selected folder");
        return;
      }

      const files = uris.map((u) => u.fsPath);
      // No focused file - show all workflows
      await WorkflowVisualizerPanel.createOrShowForFolder(context.extensionUri, folderPath, files, undefined);
    }
  );

  context.subscriptions.push(visualizeCommand);
  context.subscriptions.push(visualizeFolderCommand);

  // Watch for document changes to update visualization
  context.subscriptions.push(
    vscode.workspace.onDidSaveTextDocument((doc) => {
      if (doc.languageId === "twf") {
        WorkflowVisualizerPanel.refreshIfVisible();
      }
    })
  );

  // Watch for active editor changes to update focused file
  context.subscriptions.push(
    vscode.window.onDidChangeActiveTextEditor((editor) => {
      if (editor && editor.document.languageId === "twf") {
        WorkflowVisualizerPanel.updateFocusedFile(editor.document.uri.fsPath);
      }
    })
  );
}

function startLanguageClient(context: vscode.ExtensionContext) {
  const config = vscode.workspace.getConfiguration("twf.lsp");
  const configPath = config.get<string>("path", "");
  const command = configPath || "twf";

  const serverOptions: ServerOptions = {
    run: { command, args: ["lsp"] } as Executable,
    debug: { command, args: ["lsp"] } as Executable,
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [{ scheme: "file", language: "twf" }],
    outputChannelName: "TWF Language Server",
  };

  client = new LanguageClient(
    "twf-lsp",
    "TWF Language Server",
    serverOptions,
    clientOptions
  );

  client.start().catch((err) => {
    vscode.window.showWarningMessage(
      `Failed to start TWF language server: ${err.message}. ` +
      `Install it with: go install github.com/jmbarzee/temporal-skills/lsp/cmd/twf@latest`
    );
  });

  context.subscriptions.push({
    dispose: () => {
      if (client) {
        client.stop();
      }
    },
  });
}

export function deactivate(): Thenable<void> | undefined {
  if (client) {
    return client.stop();
  }
  return undefined;
}

/**
 * Manages workflow visualizer webview panels
 */
class WorkflowVisualizerPanel {
  public static currentPanel: WorkflowVisualizerPanel | undefined;
  public static readonly viewType = "twfVisualizer";

  private readonly _panel: vscode.WebviewPanel;
  private readonly _extensionUri: vscode.Uri;
  private _folderPath: string;
  private _files: string[];
  private _focusedFile: string | undefined;
  private _disposables: vscode.Disposable[] = [];

  /**
   * Create or show the visualizer for a single file.
   * This will parse all .twf files in the same folder for context,
   * but only show workflows from the focused file at top level.
   */
  public static async createOrShowForFile(extensionUri: vscode.Uri, filePath: string) {
    const folderPath = path.dirname(filePath);

    // Find all .twf files in the folder for context
    const pattern = new vscode.RelativePattern(folderPath, "*.twf");
    const uris = await vscode.workspace.findFiles(pattern);
    const files = uris.map((u) => u.fsPath);

    // Ensure the focused file is included
    if (!files.includes(filePath)) {
      files.push(filePath);
    }

    await WorkflowVisualizerPanel.createOrShowForFolder(extensionUri, folderPath, files, filePath);
  }

  /**
   * Create or show the visualizer for a folder with optional focused file.
   */
  public static async createOrShowForFolder(
    extensionUri: vscode.Uri,
    folderPath: string,
    files: string[],
    focusedFile: string | undefined
  ) {
    const column = vscode.ViewColumn.Beside;

    // If we already have a panel, update it (preserveFocus to not steal from editor)
    if (WorkflowVisualizerPanel.currentPanel) {
      WorkflowVisualizerPanel.currentPanel._panel.reveal(column, true);
      WorkflowVisualizerPanel.currentPanel._folderPath = folderPath;
      WorkflowVisualizerPanel.currentPanel._files = files;
      WorkflowVisualizerPanel.currentPanel._focusedFile = focusedFile;
      WorkflowVisualizerPanel.currentPanel._update();
      return;
    }

    // Create a new panel (preserveFocus to not steal from editor)
    const panel = vscode.window.createWebviewPanel(
      WorkflowVisualizerPanel.viewType,
      "TWF Visualizer",
      { viewColumn: column, preserveFocus: true },
      {
        enableScripts: true,
        retainContextWhenHidden: true,
        localResourceRoots: [
          vscode.Uri.joinPath(extensionUri, "dist", "webview"),
        ],
      }
    );

    WorkflowVisualizerPanel.currentPanel = new WorkflowVisualizerPanel(
      panel,
      extensionUri,
      folderPath,
      files,
      focusedFile
    );
  }

  public static refreshIfVisible() {
    if (WorkflowVisualizerPanel.currentPanel) {
      WorkflowVisualizerPanel.currentPanel._update();
    }
  }

  /**
   * Update the focused file and refresh the visualization.
   * Only updates if the new file is in the same folder or a .twf file.
   */
  public static async updateFocusedFile(filePath: string) {
    if (!WorkflowVisualizerPanel.currentPanel) {
      return;
    }

    const panel = WorkflowVisualizerPanel.currentPanel;
    const newFolderPath = path.dirname(filePath);

    // If the file is in a different folder, reload the folder's files
    if (newFolderPath !== panel._folderPath) {
      const pattern = new vscode.RelativePattern(newFolderPath, "*.twf");
      const uris = await vscode.workspace.findFiles(pattern);
      const files = uris.map((u) => u.fsPath);

      if (!files.includes(filePath)) {
        files.push(filePath);
      }

      panel._folderPath = newFolderPath;
      panel._files = files;
    }

    panel._focusedFile = filePath;
    panel._update();
  }

  private constructor(
    panel: vscode.WebviewPanel,
    extensionUri: vscode.Uri,
    folderPath: string,
    files: string[],
    focusedFile: string | undefined
  ) {
    this._panel = panel;
    this._extensionUri = extensionUri;
    this._folderPath = folderPath;
    this._files = files;
    this._focusedFile = focusedFile;

    // Set initial HTML content
    this._panel.webview.html = this._getHtmlForWebview();

    // Listen for when the panel is disposed
    this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

    // Handle messages from the webview
    this._panel.webview.onDidReceiveMessage(
      (message) => {
        switch (message.type) {
          case "ready":
            this._update();
            break;
          case "refocus":
            // Return focus to the last active text editor after webview interaction
            if (lastActiveTextEditor) {
              vscode.window.showTextDocument(
                lastActiveTextEditor.document,
                { viewColumn: lastActiveTextEditor.viewColumn, preserveFocus: false }
              );
            }
            break;
        }
      },
      null,
      this._disposables
    );
  }

  public dispose() {
    WorkflowVisualizerPanel.currentPanel = undefined;

    this._panel.dispose();

    while (this._disposables.length) {
      const x = this._disposables.pop();
      if (x) {
        x.dispose();
      }
    }
  }

  private async _update() {
    try {
      const ast = await this._parseFilesWithMetadata();
      this._panel.webview.postMessage({ type: "ast", data: ast });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      this._panel.webview.postMessage({ type: "error", message: errorMessage });
    }
  }

  /**
   * Parse files and add metadata for source files and focused file.
   */
  private async _parseFilesWithMetadata(): Promise<unknown> {
    const config = vscode.workspace.getConfiguration("twf.parser");
    const configPath = config.get<string>("path", "");
    const parts = ((configPath || "twf") + " parse").split(/\s+/);
    const parserCommand = parts[0];
    const baseArgs = parts.slice(1);

    if (this._files.length === 0) {
      throw new Error("No .twf files to parse");
    }

    // Parse each file individually to track source files
    const allDefinitions: unknown[] = [];
    const allErrors: { file: string; error: string; stderr?: string }[] = [];

    for (const file of this._files) {
      try {
        const { stdout, stderr } = await execFileAsync(parserCommand, [
          ...baseArgs,
          "--json",
          file,
        ]);

        if (stderr) {
          console.warn("Parser stderr:", stderr);
          allErrors.push({ file, error: `Parser warning`, stderr: stderr.trim() });
        }

        const parsed = JSON.parse(stdout) as { definitions?: unknown[] };

        // Add sourceFile to each definition
        if (parsed.definitions) {
          for (const def of parsed.definitions) {
            (def as { sourceFile?: string }).sourceFile = file;
            allDefinitions.push(def);
          }
        }
      } catch (err) {
        // Collect per-file errors instead of silently swallowing them
        const errMsg = err instanceof Error ? err.message : String(err);
        const stderr = (err as { stderr?: string }).stderr;
        allErrors.push({
          file,
          error: errMsg,
          stderr: stderr ? stderr.trim() : undefined,
        });
        console.warn(`Failed to parse ${file}:`, err);
      }
    }

    // Return combined AST with focusedFile metadata and any errors
    return {
      definitions: allDefinitions,
      errors: allErrors.length > 0 ? allErrors : undefined,
      focusedFile: this._focusedFile,
    };
  }

  private _getHtmlForWebview(): string {
    const webview = this._panel.webview;

    // Get URIs for webview resources
    const scriptUri = webview.asWebviewUri(
      vscode.Uri.joinPath(this._extensionUri, "dist", "webview", "visualizer.js")
    );
    const styleUri = webview.asWebviewUri(
      vscode.Uri.joinPath(this._extensionUri, "dist", "webview", "visualizer.css")
    );

    // Use a nonce to only allow specific scripts
    const nonce = getNonce();

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src ${webview.cspSource} 'unsafe-inline'; script-src 'nonce-${nonce}';">
    <link href="${styleUri}" rel="stylesheet">
    <title>TWF Workflow Visualizer</title>
    <style>
      html, body, #root {
        height: 100%;
        width: 100%;
        margin: 0;
        padding: 0;
        overflow: auto;
      }
    </style>
</head>
<body class="vscode-dark">
    <div id="root"></div>
    <script nonce="${nonce}" src="${scriptUri}"></script>
</body>
</html>`;
  }
}

function getNonce(): string {
  let text = "";
  const possible =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  for (let i = 0; i < 32; i++) {
    text += possible.charAt(Math.floor(Math.random() * possible.length));
  }
  return text;
}
