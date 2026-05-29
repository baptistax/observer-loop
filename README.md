# observer-loop

Small Windows Code that 'attach' to another PID until it dies

When the watched process exits, it shows a `MessageBox` with:

```text
the {process.name} killed itself...
```

Then it hops to another accessible process, prints a short trail in the terminal, and keeps running until you press `Ctrl+C`.


Included:
- explicit PID attach from the command line;
- visible terminal execution;
- short trail with `PID | process name`;
- automatic re-attach after the watched process exits;
- low-cost waiting with Windows process handles.


```bash
go build .\cmd\watch\
```

## Usage

```bash
.\ol.exe -pid 30104
...
<CTRL-C>
```

Optional title:

```bash
.\ol.exe -pid 30104 -title "Observer Loop"
```

## License

MIT
