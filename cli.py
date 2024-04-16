import click
import face_recognition
import numpy as np
import base64
import json
import os
import uuid

from sqlalchemy import create_engine, func
from database.db import Session, engine
from database.models import Person, Image, Base
from recognition.face_recognition import recognize_faces
from recognition.utils import read_image
from PIL import Image as PilImage

@click.group()
def cli():
    pass

@cli.command()
def init_db():
    Base.metadata.create_all(engine)
    click.echo("Initialized the database.")

def search_person_in_db(image_path):
    session = Session()
    unknown_image = read_image(image_path)
    unknown_face_encodings = face_recognition.face_encodings(unknown_image)

    if not unknown_face_encodings:
        return {"Found": "No face detected in the input image.", "main_image": ""}

    for unknown_face_encoding in unknown_face_encodings:
        for person in session.query(Person).all():
            known_face_encodings = [np.frombuffer(img.image, dtype=np.float64) for img in person.images]
            results = face_recognition.compare_faces(known_face_encodings, unknown_face_encoding, tolerance=0.6)
            if True in results:
                return {"Found": person.name, "main_image": person.main_image}

    return {"Found": "No match found in the database.", "main_image": ""}

@cli.command()
@click.argument('name')
@click.argument('image_paths', nargs=-1)
def add_person(name, image_paths):
    session = Session()
    images = []

    for image_path in image_paths:
        search_result = search_person_in_db(image_path)
        if search_result["Found"] == "No match found in the database.":
            image = read_image(image_path)
            face_encodings = face_recognition.face_encodings(image)
            if not face_encodings:
                error_message = f"No faces found in the image {image_path}."
                click.echo(json.dumps({"Error": error_message}))
                return

            images.append((face_encodings[0].tobytes(), image))
        else:
            error_message = f"A match for the face in the image was found in the database."
            click.echo(json.dumps({"Error": error_message}))
            return

    if images:
        person = Person(name=name)
        session.add(person)
        session.commit()

        main_image = images[0][1]
        main_image_pil = PilImage.fromarray(main_image)
        script_dir = os.path.dirname(os.path.realpath(__file__))
        main_image_dir = os.path.join(script_dir, 'database', 'main_images')
        main_image_filename = f'{uuid.uuid4()}.jpg'
        main_image_path = os.path.join(main_image_dir, main_image_filename)

        main_image_pil.save(main_image_path, format='JPEG')

        person.main_image = main_image_path
        session.commit()

        for image, _ in images:
            img = Image(person_id=person.id, image=image)
            session.add(img)
        session.commit()
        click.echo(json.dumps({"Success": f"Added {name} to the database."}))
    else:
        click.echo(json.dumps({"Error": "No valid images were provided."}))

@cli.command()
@click.argument('image_path')
def search_person(image_path):
    result = search_person_in_db(image_path)
    click.echo(json.dumps(result))

@cli.command()
@click.argument('name')
def search_name(name):
    session = Session()
    people = session.query(Person).filter(Person.name.ilike(f"%{name}%")).limit(10).all()

    if not people:
        click.echo(json.dumps({"Error": f"No person named {name} found in the database."}))
        return

    results = []
    for person in people:
        results.append({
            "name": person.name,
            "main_image": person.main_image
        })

    click.echo(json.dumps({"Success": results}))

@cli.command()
@click.argument('name')
@click.argument('image_paths', nargs=-1)
def update_person(name, image_paths):
    session = Session()
    person = session.query(Person).filter(Person.name.like(f"%{name}%")).first()

    if not person:
        click.echo(json.dumps({"Error": f"No person named {name} found in the database."}))
        return

    images = []
    for image_path in image_paths:
        image = read_image(image_path)
        face_encodings = face_recognition.face_encodings(image)
        if not face_encodings:
            click.echo(json.dumps({"Error": f"No faces found in the image {image_path}."}))
            return

        known_face_encodings = [np.frombuffer(img.image, dtype=np.float64) for img in person.images]
        results = face_recognition.compare_faces(known_face_encodings, face_encodings[0], tolerance=0.7)
        if True in results:
            images.append((face_encodings[0].tobytes(), image))

    if images:
        for image, _ in images:
            img = Image(person_id=person.id, image=image)
            session.add(img)
        session.commit()
        click.echo(json.dumps({"Success": f"Updated {person.name} in the database."}))
    else:
        click.echo(json.dumps({"Error": "No new faces found in the images or no match found in the database for the provided images."}))

if __name__ == '__main__':
    cli()