# Auto research - MKT-002

No pude ejecutar el modo autonomo con internet e IA.

Motivo: Claude CLI fallo con codigo 1: Error: Reached max turns (4)

Si usas OpenAI, necesitas `OPENAI_API_KEY`.
Si usas Claude Code, necesitas tener `claude` logueado y correr:

python .\startup_ai_lab.py auto-research --provider claude_cli --question-id MKT-002
