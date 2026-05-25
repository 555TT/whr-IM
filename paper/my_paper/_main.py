"""Combine all sections and write the docx."""
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from _gen_paper import setup_doc, ROOT
from _content_front import write_front
from _content_ch1_2 import write_ch1, write_ch2
from _content_ch3 import write_ch3
from _content_ch4 import write_ch4
from _content_ch5_end import write_ch5, write_end, write_refs


def main():
    doc = setup_doc()
    write_front(doc)
    write_ch1(doc)
    write_ch2(doc)
    write_ch3(doc)
    write_ch4(doc)
    write_ch5(doc)
    write_end(doc)
    write_refs(doc)
    out = os.path.join(ROOT, '基于Gin和WebSocket的IM即时通讯系统.docx')
    doc.save(out)
    print('Saved:', out)
    # Word count rough
    total = 0
    for p in doc.paragraphs:
        total += len([c for c in p.text if '\u4e00' <= c <= '\u9fff'])
    print('Chinese char count (approx):', total)


if __name__ == '__main__':
    main()
