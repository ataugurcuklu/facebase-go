import click
import face_recognition
import numpy as np
import base64
import json

from sqlalchemy import create_engine
from database.db import Session, engine
from database.models import Person, Image, Base
from recognition.face_recognition import recognize_faces
from recognition.utils import read_image

@click.group()
def cli():
    pass

@click.command()
def init_db():
    Base.metadata.create_all(engine)
    click.echo("Initialized the database.")

cli.add_command(init_db)

@cli.command()
@click.argument('name')
@click.argument('image_paths', nargs=-1)
def add_person(name, image_paths):
    session = Session()
    images = []
    for image_path in image_paths:
        image = read_image(image_path)
        face_encodings = face_recognition.face_encodings(image)
        if face_encodings:
            with open(image_path, 'rb') as f:
                image_data = f.read()
            images.append((face_encodings[0].tobytes(), image_data))
    if images:
        main_image = images[0][1]
        person = Person(name=name, main_image=main_image)
        session.add(person)
        session.commit()
        for image, _ in images:
            img = Image(person_id=person.id, image=image)
            session.add(img)
        session.commit()
        click.echo(f"Added {name} to the database.")
    else:
        click.echo(f"No faces found in the images.")

@cli.command()
@click.argument('name')
@click.argument('image_paths', nargs=-1)
def update_person(name, image_paths):
    session = Session()
    persons = session.query(Person).filter_by(name=name).all()
    if not persons:
        click.echo(f"No person named {name} found in the database.")
        return

    MAX_IMAGES_PER_PERSON = 10
    TOLERANCE = 0.7

    for image_path in image_paths:
        image = read_image(image_path)
        face_encodings = face_recognition.face_encodings(image)
        if face_encodings:
            for person in persons:
                known_face_encodings = [np.frombuffer(img.image, dtype=np.float64) for img in person.images]
                results = face_recognition.compare_faces(known_face_encodings, face_encodings[0], tolerance=TOLERANCE)
                if True in results:
                    if len(person.images) < MAX_IMAGES_PER_PERSON:
                        img = Image(person_id=person.id, image=face_encodings[0].tobytes())
                        session.add(img)
                        session.commit()
                        click.echo(f"Updated {name} in the database.")
                    else:
                        click.echo(f"Maximum number of images for {name} reached. No more images will be added.")
                else:
                    click.echo(f"The face in the new image does not match the existing images of {name}.")
        else:
            click.echo(f"No faces found in the image.")

@cli.command()
@click.argument('image_path')
def search_person(image_path):
    try:
        session = Session()
        unknown_image = read_image(image_path)
        unknown_face_encodings = face_recognition.face_encodings(unknown_image)

        if not unknown_face_encodings:
            click.echo(json.dumps({"Found": "No face detected in the input image.", "main_image": ""}))
            return

        for unknown_face_encoding in unknown_face_encodings:
            for person in session.query(Person).all():
                known_face_encodings = [np.frombuffer(img.image, dtype=np.float64) for img in person.images]
                results = face_recognition.compare_faces(known_face_encodings, unknown_face_encoding, tolerance=0.6)
                if True in results:
                    main_image_base64 = base64.b64encode(person.main_image).decode('utf-8')
                    click.echo(json.dumps({"Found": person.name, "main_image": "data:image/jpeg;base64," + main_image_base64}))
                    return

        click.echo(json.dumps({"Found": "No match found in the database.", "main_image": ""}))
    except Exception as e:
        click.echo(json.dumps({"Found": str(e), "main_image": ""}))

if __name__ == '__main__':
    cli()