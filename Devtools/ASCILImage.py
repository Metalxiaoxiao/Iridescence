from PIL import Image
from PyQt6.QtWidgets import QApplication, QMainWindow, QLabel, QPushButton, QVBoxLayout, QWidget, QFileDialog, QMessageBox, QLineEdit
from PyQt6.QtCore import Qt

ascii_char = list("$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/\|()1{}[]?-_+~<>i!lI;:,\"^`'. ")

# 将256灰度映射到70个字符上
def get_char(r, g, b, alpha=256):
    if alpha == 0:
        return ' '
    length = len(ascii_char)
    gray = int(0.2126 * r + 0.7152 * g + 0.0722 * b)
    unit = (256.0 + 1) / length
    return ascii_char[int(gray / unit)]

def convert_to_ascii(image_path, output_folder, width, height):
    try:
        im = Image.open(image_path)
        im = im.resize((width, height), Image.NEAREST)
        txt = ""
        for i in range(height):
            for j in range(width):
                txt += get_char(*im.getpixel((j, i)))
            txt += '\n'
        output_file = f"{output_folder}/output.txt"
        with open(output_file, 'w') as f:
            f.write(txt)
        QMessageBox.information(None, "转换成功", "ASCII艺术成功创建！")
    except Exception as e:
        QMessageBox.critical(None, "错误", str(e))

class MainWindow(QMainWindow):
    def __init__(self):
        super().__init__()
        self.setWindowTitle("ASCII艺术转换器")

        # 创建主部件和布局
        self.main_widget = QWidget()
        self.layout = QVBoxLayout(self.main_widget)

        # 创建标签
        self.image_path_label = QLabel("图像路径:")
        self.output_path_label = QLabel("输出文件夹:")
        self.width_label = QLabel("宽度:")
        self.height_label = QLabel("高度:")

        # 创建输入框
        self.image_path_entry = QLineEdit()
        self.output_path_entry = QLineEdit()
        self.width_entry = QLineEdit()
        self.height_entry = QLineEdit()

        # 创建按钮
        self.open_button = QPushButton("打开")
        self.output_folder_button = QPushButton("选择文件夹")
        self.convert_button = QPushButton("转换")

        # 连接按钮的点击事件到相应的函数
        self.open_button.clicked.connect(self.open_file)
        self.output_folder_button.clicked.connect(self.choose_output_folder)
        self.convert_button.clicked.connect(self.convert)

        # 设置标签的对齐方式
        self.image_path_label.setAlignment(Qt.AlignmentFlag.AlignLeft)
        self.output_path_label.setAlignment(Qt.AlignmentFlag.AlignLeft)
        self.width_label.setAlignment(Qt.AlignmentFlag.AlignLeft)
        self.height_label.setAlignment(Qt.AlignmentFlag.AlignLeft)

        # 将部件添加到布局
        self.layout.addWidget(self.image_path_label)
        self.layout.addWidget(self.image_path_entry)
        self.layout.addWidget(self.open_button)
        self.layout.addWidget(self.output_path_label)
        self.layout.addWidget(self.output_path_entry)
        self.layout.addWidget(self.output_folder_button)
        self.layout.addWidget(self.width_label)
        self.layout.addWidget(self.width_entry)
        self.layout.addWidget(self.height_label)
        self.layout.addWidget(self.height_entry)
        self.layout.addWidget(self.convert_button)

        # 设置布局为主部件的布局
        self.main_widget.setLayout(self.layout)

        # 将主部件设置为主窗口的中央部件
        self.setCentralWidget(self.main_widget)

    def open_file(self):
        file_dialog = QFileDialog()
        file_path, _ = file_dialog.getOpenFileName(self, "打开图像", "", "图像文件 (*.png *.jpg *.jpeg)")
        if file_path:
            self.image_path_entry.setText(file_path)

    def choose_output_folder(self):
        dialog = QFileDialog()
        folder = dialog.getExistingDirectory(self, "选择输出文件夹")
        if folder:
            self.output_path_entry.setText(folder)

    def convert(self):
        image_path = self.image_path_entry.text()
        output_folder = self.output_path_entry.text()
        width_text = self.width_entry.text()
        height_text = self.height_entry.text()

        # 检查宽度和高度是否为空
        if not width_text or not height_text:
            QMessageBox.critical(None, "错误", "请输入宽度和高度。")
            return

        # 将宽度和高度转换为整数
        try:
            width = int(width_text)
            height = int(height_text)
        except ValueError:
            QMessageBox.critical(None, "错误", "无效的宽度或高度。")
            return

        convert_to_ascii(image_path, output_folder, width, height)


app = QApplication([])
app.setStyle("Fusion")  # 设置应用程序风格为Fusion（用于更改语言）
window = MainWindow()
window.show()
app.exec()
