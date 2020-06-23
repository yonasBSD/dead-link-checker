'''This module contains the models used in DeLiC'''
# pylint: disable=no-name-in-module

from typing import List

from pydantic import BaseModel


class Link(BaseModel):
    '''Represents a single link with its attributes'''
    url: str
    page: str  # Page containing the url
    status: str = None


class SiteResultDetails(BaseModel):
    '''Details of the results of a single site'''
    broken: List[Link]


class SiteResultSummary(BaseModel):
    '''Summary of the results of a single site'''
    urls_checked: int
    urls_broken: int


class SiteResult(BaseModel):
    '''Results of a single site'''
    site: str
    summary: SiteResultSummary
    details: SiteResultDetails


class SiteResultList(BaseModel):
    '''List of site results'''
    __root__: List[SiteResult] = []
