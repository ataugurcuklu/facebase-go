from sqlalchemy import Column, Integer, String, ForeignKey, LargeBinary
from sqlalchemy.orm import relationship
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()

class Person(Base):
    __tablename__ = 'people'

    id = Column(Integer, primary_key=True)
    name = Column(String)
    main_image = Column(String)

class Image(Base):
    __tablename__ = 'images'

    id = Column(Integer, primary_key=True)
    person_id = Column(Integer, ForeignKey('people.id'))
    image = Column(LargeBinary)
    person = relationship("Person", back_populates="images")

Person.images = relationship("Image", order_by=Image.id, back_populates="person")