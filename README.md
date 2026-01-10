# sluhach

Simple speech‑to‑text CLI tool based on [Vosk](https://alphacephei.com/vosk/).

`sluhach` records audio from your default microphone, recognizes speech locally
using Vosk models, copies recognized text to the clipboard, and shows desktop
notifications about recording status. It also includes simple commands to
manage local and remote models.

## Features

- Record from the default microphone and convert speech to text
- Fully local recognition using Vosk models (no external API calls)
- Automatic stop after a period of silence
- Print recognized text to the terminal
- Optional copying of recognized text to the system clipboard
- Desktop notifications on recording start and finish
- Manage Vosk models (list, load/download, remove, show available)

## Basic Usage

Run `sluhach` without arguments to see the built‑in help:

```bash
sluhach --help
```

The main functionality is exposed via two top‑level commands:

- `sluhach reco` – record from the microphone and recognize speech
- `sluhach model` – manage speech recognition models

---

## `reco` – record and recognize speech

The `reco` command records from the default microphone and recognizes speech
using a selected Vosk model.

By default it:

- Uses the `vosk-model-small-ru-0.22` model
- Stops recording after several seconds of silence (configurable)
- Prints the recognized text to the terminal
- Copies the recognized text to the clipboard

Aliases:

- `sluhach reco`
- `sluhach r`

### Examples

Record from the microphone, recognize speech and copy text to the clipboard:

```bash
sluhach reco
```

Use an English model:

```bash
sluhach reco -m vosk-model-en-us-0.22
```

Wait up to 8 seconds of silence before stopping recording:

```bash
sluhach reco -w 8
```

Print the recognized text only (do not copy to clipboard):

```bash
sluhach reco --no-paste
```

### Flags

- `-m, --model string` – model name to use
  - Default: `vosk-model-small-ru-0.22`
- `-w, --wait int` – seconds of silence before recording stops
  - Default: `5`
- `--no-paste` – do not copy recognized text to the clipboard

During recording, `sluhach` prints a message indicating that it is listening
and waiting for silence to stop. When recognition finishes, you will see the
recognized text in the terminal and, unless `--no-paste` is used, a message
confirming that the text was copied to the clipboard. Desktop notifications
are also shown for recording start and finish.

---

## `model` – manage speech recognition models

The `model` command groups subcommands for working with Vosk models.

Aliases:

- `sluhach model`
- `sluhach m`

Run without subcommands to see help:

```bash
sluhach model --help
```

### `model list` – list installed models

Show models that are already downloaded and available locally.

```bash
sluhach model list
```

This inspects the local models directory and prints a table with:

- `#` – index number
- `Name` – model name
- `Path` – filesystem path to the model

### `model load` – download a model

Download and unpack a Vosk model by name into the local models directory.

```bash
sluhach model load [name]

# Examples
sluhach model load vosk-model-small-ru-0.22
sluhach model load vosk-model-en-us-0.22
```

Notes:

- The model archive is fetched from a configured base URL and extracted into
  the local models directory.
- If the model is already present, the command returns an error.

### `model remove` – remove a model

Remove a previously downloaded model from the local models directory.

```bash
sluhach model remove [name]

# Example
sluhach model remove vosk-model-small-ru-0.22
```

Only local files are deleted; no remote resources are affected.

### `model avail` – list models available for download

Show models that are available for download from the remote repository.

```bash
sluhach model avail
```

The output is a table with:

- `#` – index number
- `Lang` – language code
- `Name` – model name
- `Size` – approximate size of the model
- `Desc` – short description

Use the `Name` from this list with `sluhach model load` to download a
particular model.

---

## Notes

- All recognition runs locally using Vosk models; an internet connection is
  only required when downloading new models (depending on your configuration).
- Clipboard and notifications depend on the underlying OS support and may
  behave slightly differently across platforms.
