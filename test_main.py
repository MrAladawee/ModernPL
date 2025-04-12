import io
import os
import time
import numpy as np
import cv2
import pytest
from PIL import Image
from fastapi.testclient import TestClient
from pathlib import Path

# Импортируем приложение и data_dir из основного модуля.
# Предполагается, что основной код находится в файле main.py.
from main import app, data_dir, generate_video_thumbnail

client = TestClient(app)


@pytest.fixture(autouse=True)
def temp_upload_dir(tmp_path, monkeypatch):
    # Подменяем директорию для загрузки на временную, чтобы тесты не оставляли файлы в uploads/
    temp_dir = tmp_path / "uploads"
    temp_dir.mkdir()
    monkeypatch.setattr("main.data_dir", temp_dir)
    yield
    # После теста временная директория будет удалена вместе с tmp_path


def create_test_image_bytes(format="PNG"):
    # Создаём простое изображение в памяти
    img = Image.new("RGB", (100, 100), color="red")
    img_byte_arr = io.BytesIO()
    img.save(img_byte_arr, format=format)
    img_byte_arr.seek(0)
    return img_byte_arr


def create_test_video_bytes():
    # Создаём фиктивное "видео" как бинарный поток.
    # Так как проверка в PUT основана только на mime-типе (по расширению), можно отправить любые байты.
    return io.BytesIO(b"fake video data")


# Для тестов видео будем подменять поведение cv2.VideoCapture
class FakeVideoCapture:
    def __init__(self, filename):
        self.filename = filename
        self.called = False

    def read(self):
        # Возвращаем dummy-кадр: массив numpy с нужными размерами
        self.called = True
        # Создадим "кадр" размером 200x200 с 3 каналами
        frame = np.full((200, 200, 3), 255, dtype=np.uint8)
        return True, frame

    def release(self):
        pass


@pytest.fixture
def fake_video(monkeypatch):
    monkeypatch.setattr(cv2, "VideoCapture", lambda filename: FakeVideoCapture(filename))


### Тесты для метода загрузки (PUT /api/upload)

def test_upload_valid_image():
    # Создаём изображение и отправляем запрос
    image_bytes = create_test_image_bytes()
    files = {"file": ("test.png", image_bytes, "image/png")}
    response = client.put("/api/upload", files=files)
    assert response.status_code == 200
    data = response.json()
    assert "uuid" in data
    assert data["название"] == "test.png"
    assert data["расширение"] == "img"
    # Проверяем, что вес и временные метки присутствуют
    assert isinstance(data["вес"], int)
    assert isinstance(data["когда добавлен"], (int, float))
    assert isinstance(data["последнее изменение"], (int, float))

def test_upload_valid_video(fake_video):
    # Создаём фиктивное видео
    video_bytes = create_test_video_bytes()
    files = {"file": ("test.mp4", video_bytes, "video/mp4")}
    response = client.put("/api/upload", files=files)
    assert response.status_code == 200
    data = response.json()
    assert "uuid" in data
    assert data["название"] == "test.mp4"
    assert data["расширение"] == "видео"

def test_upload_corrupted_video():
    import io
    # Создаем "битый" видеофайл: данные, которые не распознаются OpenCV как валидное видео
    broken_video = io.BytesIO(b"this is not a valid video content")
    files = {"file": ("corrupted.mp4", broken_video, "video/mp4")}
    response = client.put("/api/upload", files=files)
    assert response.status_code == 400
    data = response.json()
    assert "corrupted" in data["detail"]

def test_upload_invalid_file():
    # Отправляем файл с неподходящим MIME типом (.txt)
    txt_bytes = io.BytesIO(b"Hello, world!")
    files = {"file": ("test.txt", txt_bytes, "text/plain")}
    response = client.put("/api/upload", files=files)
    assert response.status_code == 400
    data = response.json()
    assert "Only images and videos are allowed" in data["detail"]


### Тесты для метода получения (GET /api/{uuid})

def test_get_file_image():
    # Сначала загружаем изображение
    image_bytes = create_test_image_bytes()
    files = {"file": ("image.png", image_bytes, "image/png")}
    response = client.put("/api/upload", files=files)
    data = response.json()
    file_uuid = data["uuid"]

    # Получаем файл по UUID
    response = client.get(f"/api/{file_uuid}")
    assert response.status_code == 200
    # Проверяем, что Content-Disposition содержит имя файла
    assert "attachment; filename=" in response.headers.get("Content-Disposition", "")


def test_get_file_image_thumbnail():
    # Загружаем изображение
    image_bytes = create_test_image_bytes()
    files = {"file": ("thumb_test.png", image_bytes, "image/png")}
    response = client.put("/api/upload", files=files)
    data = response.json()
    file_uuid = data["uuid"]

    # Запрос с параметрами ширины и высоты
    response = client.get(f"/api/{file_uuid}?width=50&height=50")
    assert response.status_code == 200
    # Можно проверить, что возвращается изображение (например, Content-Type image)
    content_type = response.headers.get("content-type")
    assert content_type.startswith("image/")


def test_get_file_video_thumbnail(fake_video):
    # Загружаем фиктивное видео
    video_bytes = create_test_video_bytes()
    files = {"file": ("video.mp4", video_bytes, "video/mp4")}
    response = client.put("/api/upload", files=files)
    data = response.json()
    file_uuid = data["uuid"]

    # Запрос с параметрами для получения thumbnail видео.
    response = client.get(f"/api/{file_uuid}?width=100&height=100")
    # Так как FakeVideoCapture возвращает корректный dummy-кадр,
    # ожидаем успешный ответ с Content-Type image/jpeg (так как сохраняется в jpg)
    assert response.status_code == 200
    content_type = response.headers.get("content-type")
    assert content_type.startswith("image/")


def test_get_nonexistent_file():
    # Запрос по несуществующему UUID
    response = client.get("/api/nonexistent_uuid")
    assert response.status_code == 404
    data = response.json()
    assert "File not found" in data["detail"]


### Тесты для метода удаления (DELETE /api/{uuid})

def test_delete_file():
    # Загружаем изображение
    image_bytes = create_test_image_bytes()
    files = {"file": ("delete_test.png", image_bytes, "image/png")}
    response = client.put("/api/upload", files=files)
    data = response.json()
    file_uuid = data["uuid"]

    # Удаляем файл
    response = client.delete(f"/api/{file_uuid}")
    assert response.status_code == 200
    del_data = response.json()
    # Проверяем, что в ответе присутствуют данные удалённого файла
    assert "deleted_file" in del_data
    assert del_data["deleted_file"]["uuid"] == file_uuid

    # Повторный запрос удаления должен вернуть ошибку 404
    response = client.delete(f"/api/{file_uuid}")
    assert response.status_code == 404


def test_delete_nonexistent_file():
    # Попытка удаления несуществующего файла
    response = client.delete("/api/nonexistent_uuid")
    assert response.status_code == 404
    data = response.json()
    assert "File not found" in data["detail"]
