from typing import TYPE_CHECKING, Any, Dict, List, Type, TypeVar, Union, cast
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.tool_attributes import ToolAttributes


T = TypeVar("T", bound="Tool")


@_attrs_define
class Tool:
    """
    Attributes:
        id (Union[Unset, UUID]):
        run_id (Union[Unset, UUID]):
        name (Union[Unset, str]):
        description (Union[Unset, str]):
        attributes (Union[Unset, ToolAttributes]):
        ignored_attributes (Union[Unset, List[str]]):
    """

    id: Union[Unset, UUID] = UNSET
    run_id: Union[Unset, UUID] = UNSET
    name: Union[Unset, str] = UNSET
    description: Union[Unset, str] = UNSET
    attributes: Union[Unset, "ToolAttributes"] = UNSET
    ignored_attributes: Union[Unset, List[str]] = UNSET
    additional_properties: Dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        id: Union[Unset, str] = UNSET
        if not isinstance(self.id, Unset):
            id = str(self.id)

        run_id: Union[Unset, str] = UNSET
        if not isinstance(self.run_id, Unset):
            run_id = str(self.run_id)

        name = self.name

        description = self.description

        attributes: Union[Unset, Dict[str, Any]] = UNSET
        if not isinstance(self.attributes, Unset):
            attributes = self.attributes.to_dict()

        ignored_attributes: Union[Unset, List[str]] = UNSET
        if not isinstance(self.ignored_attributes, Unset):
            ignored_attributes = self.ignored_attributes

        field_dict: Dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if id is not UNSET:
            field_dict["id"] = id
        if run_id is not UNSET:
            field_dict["run_id"] = run_id
        if name is not UNSET:
            field_dict["name"] = name
        if description is not UNSET:
            field_dict["description"] = description
        if attributes is not UNSET:
            field_dict["attributes"] = attributes
        if ignored_attributes is not UNSET:
            field_dict["ignored_attributes"] = ignored_attributes

        return field_dict

    @classmethod
    def from_dict(cls: Type[T], src_dict: Dict[str, Any]) -> T:
        from ..models.tool_attributes import ToolAttributes

        d = src_dict.copy()
        _id = d.pop("id", UNSET)
        id: Union[Unset, UUID]
        if isinstance(_id, Unset):
            id = UNSET
        else:
            id = UUID(_id)

        _run_id = d.pop("run_id", UNSET)
        run_id: Union[Unset, UUID]
        if isinstance(_run_id, Unset):
            run_id = UNSET
        else:
            run_id = UUID(_run_id)

        name = d.pop("name", UNSET)

        description = d.pop("description", UNSET)

        _attributes = d.pop("attributes", UNSET)
        attributes: Union[Unset, ToolAttributes]
        if isinstance(_attributes, Unset):
            attributes = UNSET
        else:
            attributes = ToolAttributes.from_dict(_attributes)

        ignored_attributes = cast(List[str], d.pop("ignored_attributes", UNSET))

        tool = cls(
            id=id,
            run_id=run_id,
            name=name,
            description=description,
            attributes=attributes,
            ignored_attributes=ignored_attributes,
        )

        tool.additional_properties = d
        return tool

    @property
    def additional_keys(self) -> List[str]:
        return list(self.additional_properties.keys())

    def __getitem__(self, key: str) -> Any:
        return self.additional_properties[key]

    def __setitem__(self, key: str, value: Any) -> None:
        self.additional_properties[key] = value

    def __delitem__(self, key: str) -> None:
        del self.additional_properties[key]

    def __contains__(self, key: str) -> bool:
        return key in self.additional_properties
