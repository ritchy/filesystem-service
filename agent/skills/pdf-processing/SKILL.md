<!--
from -> https://agentskills.io/

dir structure for a skill:
>my-skill/
├── SKILL.md          # Required: instructions + metadata
├── scripts/          # Optional: executable code
├── references/       # Optional: documentation
└── assets/           # Optional: templates, resources

This skill follows the standard skill file convention with YAML frontmatter (`name`, `description`) and structured instructions covering when to use it, how to extract PDF text with `pdfplumber`, and how to fill PDF forms with `pdfrw`. 

Skills like this act as persistent, reusable context so agents can automatically apply the right tools and approach whenever a PDF-related task comes up in this project.
-->

---
name: pdf-processing
description: Extract PDF text, fill forms, merge files. Use when handling PDFs.
---

# PDF Processing

## When to use this skill
Use this skill when the user needs to work with PDF files...

## How to extract text
1. Use pdfplumber for text extraction...

## How to fill forms

1. Use pdfrw to read the PDF template...
2. Update the form fields with the desired values...
3. Save the filled PDF to a new file... 
