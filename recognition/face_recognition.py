import face_recognition

def recognize_faces(known_faces, unknown_face):
    unknown_face_encoding = face_recognition.face_encodings(unknown_face)[0]
    results = face_recognition.compare_faces(known_faces, unknown_face_encoding)
    return results