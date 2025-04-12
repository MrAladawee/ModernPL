from fastapi import FastAPI, UploadFile, File, HTTPException
from fastapi.responses import FileResponse, JSONResponse
from uuid import uuid4
from pathlib import Path
import shutil
from PIL import Image
import cv2
import mimetypes
import os
import time

data_dir = Path("uploads")
data_dir.mkdir(exist_ok=True)

app = FastAPI()


def is_valid_media_type(file: UploadFile):
    mime_type, _ = mimetypes.guess_type(file.filename)
    return mime_type and (mime_type.startswith("image/") or mime_type.startswith("video/"))


def get_file_info(file_path: Path, original_filename: str = None, file_uuid: str = None):
    # Получаем информацию о файле
    stat = file_path.stat()
    mime_type, _ = mimetypes.guess_type(file_path)
    if mime_type and mime_type.startswith("image/"):
        ext_type = "img"
    elif mime_type and mime_type.startswith("video/"):
        ext_type = "видео"
    else:
        ext_type = "неизвестно"

    info = {
        "название": original_filename if original_filename else file_path.name,
        "вес": stat.st_size,
        "когда добавлен": stat.st_ctime,
        "последнее изменение": stat.st_mtime,
        "расширение": ext_type,
        "uuid": file_uuid if file_uuid else file_path.stem.split("_thumb")[0]
    }
    return info


@app.put("/api/upload")
async def upload_file(file: UploadFile = File(...)):
    if not is_valid_media_type(file):
        raise HTTPException(status_code=400, detail="Only images and videos are allowed")

    file_uuid = str(uuid4())
    file_ext = Path(file.filename).suffix
    file_path = data_dir / f"{file_uuid}{file_ext}"

    with file_path.open("wb") as buffer:
        shutil.copyfileobj(file.file, buffer)

    # Проверка целостности видео: пытаемся прочитать первый кадр
    mime_type, _ = mimetypes.guess_type(file.filename)
    if mime_type and mime_type.startswith("video/"):
        cap = cv2.VideoCapture(str(file_path))
        success, _ = cap.read()
        cap.release()
        if not success:
            os.remove(file_path)
            raise HTTPException(status_code=400, detail="Uploaded video file is corrupted or has broken bits")

    file_info = get_file_info(file_path, original_filename=file.filename, file_uuid=file_uuid)
    return JSONResponse(content=file_info)


@app.get("/api/{file_uuid}")
async def get_file(file_uuid: str, width: int = None, height: int = None):
    # Ищем файл по uuid (учитывая, что расширение может быть разным)
    files = list(data_dir.glob(f"{file_uuid}.*"))
    if not files:
        raise HTTPException(status_code=404, detail="File not found")
    file_path = files[0]
    mime_type, _ = mimetypes.guess_type(file_path)

    if width and height:
        if mime_type and mime_type.startswith("image/"):
            return generate_image_thumbnail(file_path, width, height)
        elif mime_type and mime_type.startswith("video/"):
            return generate_video_thumbnail(file_path, width, height)

    return FileResponse(file_path, headers={"Content-Disposition": f"attachment; filename={file_path.name}"})


def generate_image_thumbnail(image_path: Path, width: int, height: int):
    img = Image.open(image_path)
    img.thumbnail((width, height))
    thumb_path = image_path.with_name(f"{image_path.stem}_thumb{image_path.suffix}")
    img.save(thumb_path)
    return FileResponse(thumb_path)


def generate_video_thumbnail(video_path: Path, width: int, height: int):
    cap = cv2.VideoCapture(str(video_path))
    success, frame = cap.read()
    cap.release()

    if not success:
        raise HTTPException(status_code=500, detail="Failed to generate video thumbnail")

    frame = cv2.resize(frame, (width, height))
    thumb_path = video_path.with_name(f"{video_path.stem}_thumb.jpg")
    cv2.imwrite(str(thumb_path), frame)
    return FileResponse(thumb_path)


@app.delete("/api/{file_uuid}")
async def delete_file(file_uuid: str):
    files = list(data_dir.glob(f"{file_uuid}.*"))
    if not files:
        raise HTTPException(status_code=404, detail="File not found")
    file_path = files[0]
    # Получаем информацию до удаления
    file_info = get_file_info(file_path, file_uuid=file_uuid)
    try:
        os.remove(file_path)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error deleting file: {str(e)}")
    return JSONResponse(content={"deleted_file": file_info})
