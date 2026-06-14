from docx import Document
import os

# 找当前目录 docx
for file in os.listdir():
    if file.endswith(".docx"):
        docx_file = file
        break

doc = Document(docx_file)

with open("thesis.txt", "w", encoding="utf-8") as f:
    for para in doc.paragraphs:
        text = para.text.strip()
        if text:
            f.write(text + "\n")

print("转换完成 -> thesis.txt")